package main

import (
	"context"
	"crypto/subtle"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"resumetake/database"
	"resumetake/handlers"
	"resumetake/middleware"
	"resumetake/models"
	"resumetake/services"
	"resumetake/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "/app/data/resumetake.db"
	}

	jsonPath := os.Getenv("USER_DATA_FILE")
	if jsonPath == "" {
		jsonPath = "/app/data/users.json"
	}

	// R56b-B3: validate ADMIN_TOKEN BEFORE allocating any resources.
	// The prior placement (after db/userStore/handlers init, with
	// defer db.Close etc. already registered) used log.Fatalf, which
	// calls os.Exit and skips ALL defers — leaking the SQLite handle,
	// the verification store, and the payment handler on misconfig.
	// Failing fast here keeps the startup path leak-free.
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		log.Fatalf("ADMIN_TOKEN environment variable is required")
	}
	if len(adminToken) < 32 {
		log.Fatalf("ADMIN_TOKEN must be at least 32 characters long (current: %d)", len(adminToken))
	}

	middleware.InitSentry()

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("[DB] Failed to initialize database: %v", err)
	}
	defer db.Close()

	userStore := models.NewUserStore()
	verificationStore := models.NewVerificationStore()
	resumeStore := models.NewStore()
	defer verificationStore.Close()

	users, err := db.GetAllUsers()
	if err != nil {
		log.Printf("[DB] Failed to load users from SQLite: %v", err)
	} else {
		userMap := make(map[string]*models.User, len(users))
		for _, u := range users {
			userMap[u.Email] = u
		}
		userStore.Load(userMap)
		log.Printf("[DB] Loaded %d users from SQLite", len(users))
	}

	if userStore.Count() == 0 {
		persistence := store.NewUserPersistence(jsonPath)
		jsonUsers, err := persistence.Load()
		if err == nil && len(jsonUsers) > 0 {
			log.Printf("[DB] Migrating %d users from JSON to SQLite", len(jsonUsers))
			if err := db.MigrateFromJSON(jsonUsers); err != nil {
				log.Printf("[DB] Failed to migrate users: %v", err)
			} else {
				userStore.Load(jsonUsers)
				log.Printf("[DB] Migration complete")
			}
		}
	}

	resumes, err := db.GetAllResumes()
	if err != nil {
		log.Printf("[DB] Failed to load resumes from SQLite: %v", err)
	} else {
		for _, r := range resumes {
			resumeStore.Save(r)
		}
		log.Printf("[DB] Loaded %d resumes from SQLite", len(resumes))
	}

	dbPersistence := &DatabasePersistence{db: db}

	authHandler := handlers.NewAuthHandler(userStore, verificationStore, dbPersistence)
	resumeHandler := handlers.NewResumeHandler(resumeStore, userStore, db)
	jobHandler := handlers.NewJobHandler()
	paymentHandler := handlers.NewPaymentHandler(userStore, dbPersistence)
	defer paymentHandler.Close()
	productHandler := handlers.NewProductHandler(userStore, dbPersistence)

	app := fiber.New(fiber.Config{
		AppName:       "ResumeTake API v2.1",
		BodyLimit:     3 * 1024 * 1024,
		ServerHeader:  "ResumeTake",
		StrictRouting: true,
		CaseSensitive: true,
		IdleTimeout:   30 * time.Second,
		ReadTimeout:   20 * time.Second,  // Slowloris 防御：限制读取请求头+体的最长时间
		WriteTimeout:  300 * time.Second, // SSE 流式响应需要足够长的写入时间（AI 生成可达数分钟）
		// Trust nginx + Docker bridge gateways as proxies so c.IP() reads the
		// real client IP from X-Forwarded-For. Backend runs in host network
		// mode; nginx connects via Docker bridge (172.16.0.0/12) or localhost.
		TrustedProxies: []string{"127.0.0.1", "::1", "172.16.0.0/12", "192.168.0.0/16"},
		ProxyHeader:    fiber.HeaderXForwardedFor,
	})

	app.Use(recover.New())
	// R45-B2: requestid must run BEFORE Sentry so that Sentry's ConfigureScope
	// can read c.Locals("requestid") and attach it as a tag — enabling
	// correlation between Sentry events and structured access logs.
	app.Use(requestid.New())
	app.Use(middleware.Sentry())
	app.Use(middleware.StructuredLog())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
		// SSE 流式端点不能被压缩，否则会缓冲输出破坏实时性。
		// R31-5: use suffix matching so all streaming endpoints (optimize-stream,
		// linkedin-stream, interview-stream, etc.) are covered, not just one
		// hardcoded path.
		Next: func(c *fiber.Ctx) bool {
			return strings.HasSuffix(c.Path(), "-stream")
		},
	}))
	app.Use(helmet.New(helmet.Config{
		XSSProtection:       "0", // deprecated; "1; mode=block" can introduce XSS. Match nginx R4-7 setting.
		ContentTypeNosniff:  "nosniff",
		XFrameOptions:       "SAMEORIGIN",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		HSTSMaxAge:          31536000, // HSTS includes subdomains by default (HSTSExcludeSubdomains=false)
		ContentSecurityPolicy: "default-src 'none'; frame-ancestors 'none'",
	}))
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "https://resume.takee.top"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowMethods: "GET,POST,DELETE,OPTIONS",
		MaxAge:       86400,
	}))
	// R45-B3: Metrics must run BEFORE the global limiter so that rate-limited
	// (429) requests are still counted in metrics. Previously, the limiter
	// short-circuited the chain before Metrics ran, making 429s invisible
	// in /metrics — ops had to check nginx logs instead.
	app.Use(middleware.Metrics())
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Next: func(c *fiber.Ctx) bool {
			// Skip global limiter for paths with their own rate limiters
			// (webhook) or that must always succeed (health).
			path := c.Path()
			return path == "/api/health" || path == "/api/v1/paypal-webhook"
		},
	}))

	app.Get("/api/health", resumeHandler.Health)
	if os.Getenv("ENABLE_API_DOCS") == "true" {
		app.Get("/api/v1/docs", handlers.GetAPIDoc)
	}
	// Strict rate limiter for admin endpoints — limits brute-force attempts
	// on ADMIN_TOKEN beyond the global 100/min limiter. Applied to /metrics
	// as well since it shares ADMIN_TOKEN auth and exposes runtime internals
	// (memory, goroutine count, latency distribution).
	// R50-B1: each admin endpoint gets its own limiter instance. Previously
	// a single adminLimiter was shared across /metrics, /api/admin/users,
	// and /api/admin/health — a health check probe would consume the 5/min
	// quota of /api/admin/users, blocking admin queries.
	newAdminLimiter := func() fiber.Handler {
		return limiter.New(limiter.Config{
			Max:        5,
			Expiration: time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
		})
	}
	app.Get("/metrics", newAdminLimiter(), func(c *fiber.Ctx) error {
		token := middleware.ExtractBearerToken(c)
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) != 1 {
			log.Printf("[ADMIN] failed auth attempt on /metrics from IP %s", c.IP())
			return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
		}
		return middleware.MetricsHandler()(c)
	})
	app.Get("/api/config", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"paypal_client_id": os.Getenv("PAYPAL_CLIENT_ID"),
		})
	})

	app.Get("/api/admin/users", newAdminLimiter(), func(c *fiber.Ctx) error {
		token := middleware.ExtractBearerToken(c)
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) != 1 {
			log.Printf("[ADMIN] failed auth attempt from IP %s", c.IP())
			return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
		}
		// Pagination — without this, a 10,000-user database loads every
		// user into memory, clones each (GetAll holds RLock), serializes
		// all into one JSON response, and blocks the read lock for the
		// entire duration. Defaults to page 1, 50 per page; max 200.
		page := max(1, c.QueryInt("page", 1))
		perPage := c.QueryInt("per_page", 50)
		if perPage < 1 {
			perPage = 50
		}
		if perPage > 200 {
			perPage = 200
		}
		userMap := userStore.GetAll()
		// Use typed counters instead of fiber.Map type assertions — the
		// unchecked .(int) assertions would panic if stats init ever changes.
		proCount, enterpriseCount, freeCount, totalUsage := 0, 0, 0, 0
		// Collect into a slice sorted by email for stable pagination.
		users := make([]*models.User, 0, len(userMap))
		for _, u := range userMap {
			plan := u.Plan
			if plan == "" {
				plan = "free"
			}
			switch plan {
			case "pro":
				proCount++
			case "enterprise":
				enterpriseCount++
			default:
				freeCount++
			}
			totalUsage += u.UsageCount
			users = append(users, u)
		}
		sort.Slice(users, func(i, j int) bool { return users[i].Email < users[j].Email })
		// Apply pagination to the user list (stats reflect ALL users).
		total := len(users)
		start := (page - 1) * perPage
		if start < 0 { // R58-B1: prevent integer overflow → negative slice index panic
			start = 0
		}
		if start > total {
			start = total
		}
		end := start + perPage
		if end > total {
			end = total
		}
		userList := make([]fiber.Map, 0, len(users[start:end]))
		for _, u := range users[start:end] {
			plan := u.Plan
			if plan == "" {
				plan = "free"
			}
			userList = append(userList, fiber.Map{
				"id": u.ID, "email": u.Email, "name": u.Name,
				"plan": plan, "usage_count": u.UsageCount,
				"created_at": u.CreatedAt.Format(time.RFC3339),
			})
		}
		stats := fiber.Map{
			"total":       total,
			"pro":         proCount,
			"enterprise":  enterpriseCount,
			"free":        freeCount,
			"total_usage": totalUsage,
		}
		return c.JSON(fiber.Map{
			"success": true,
			"stats":   stats,
			"users":   userList,
			"pagination": fiber.Map{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": (total + perPage - 1) / perPage,
			},
		})
	})

	app.Get("/api/admin/health", newAdminLimiter(), func(c *fiber.Ctx) error {
		token := middleware.ExtractBearerToken(c)
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) != 1 {
			return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Authentication required"})
		}
		return resumeHandler.AdminHealth(c)
	})

	v1 := app.Group("/api/v1")

	v1.Post("/resumes", middleware.AuthMiddleware(userStore), resumeHandler.Create)
	v1.Get("/resumes", middleware.AuthMiddleware(userStore), resumeHandler.List)
	v1.Get("/resumes/:id", middleware.AuthMiddleware(userStore), resumeHandler.Get)
	v1.Delete("/resumes/:id", middleware.AuthMiddleware(userStore), resumeHandler.Delete)
	v1.Post("/upload", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), resumeHandler.Upload)
	v1.Post("/optimize", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), resumeHandler.Optimize)
	v1.Post("/perspective", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), resumeHandler.Perspective)
	v1.Post("/generate-resume", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), resumeHandler.GenerateResume)
	v1.Post("/optimize-stream", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), resumeHandler.OptimizeStream)

	v1.Post("/scrape-job", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), jobHandler.ScrapeJob)
	v1.Get("/jobs", jobHandler.GetJobs)
	v1.Get("/jobs/:id", jobHandler.GetJob)

	v1.Post("/auth/send-code", limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), authHandler.SendCode)
	v1.Post("/auth/verify-code", limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), authHandler.VerifyCode)
	v1.Post("/auth/verify-login", limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), authHandler.VerifyLogin)
	v1.Post("/auth/register", limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), authHandler.Register)
	v1.Post("/auth/login", limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), authHandler.Login)
	v1.Get("/auth/me", middleware.AuthMiddleware(userStore), authHandler.Me)
	v1.Post("/auth/logout", middleware.AuthMiddleware(userStore), authHandler.Logout)

	v1.Get("/pricing", paymentHandler.GetPricing)
	v1.Get("/templates", paymentHandler.GetTemplates)
	// Rate-limit PayPal order creation/capture endpoints — each call hits
	// PayPal's API, and dangling orders from abuse can trigger PayPal's own
	// rate limits or pollute the merchant dashboard.
	paypalLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
	v1.Post("/create-paypal-order", middleware.AuthMiddleware(userStore), paypalLimiter, paymentHandler.CreatePayPalOrder)
	v1.Post("/capture-paypal-order", middleware.AuthMiddleware(userStore), paypalLimiter, paymentHandler.CapturePayPalOrder)
	// Rate-limit webhook endpoint — each call triggers PayPal's
	// verify-webhook-signature API (expensive external call). Without a
	// dedicated limiter, attackers can send 100 forged webhooks/min (global
	// limit) and trigger PayPal's own rate limits, blocking legitimate
	// webhooks. 30/min is generous for PayPal's retry behavior.
	webhookLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
	v1.Post("/paypal-webhook", webhookLimiter, paymentHandler.PayPalWebhook)

	v1.Get("/products", productHandler.GetProducts)
	v1.Post("/purchase-product", middleware.AuthMiddleware(userStore), paypalLimiter, productHandler.PurchaseProduct)
	v1.Post("/purchase-template", middleware.AuthMiddleware(userStore), paypalLimiter, productHandler.PurchaseTemplate)
	v1.Post("/capture-template-order", middleware.AuthMiddleware(userStore), paypalLimiter, productHandler.CaptureTemplateOrder)
	v1.Post("/cover-letter", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), productHandler.GenerateCoverLetter)
	v1.Post("/linkedin-optimize", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), productHandler.OptimizeLinkedIn)
	v1.Post("/interview-practice", middleware.AuthMiddleware(userStore), limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), productHandler.PracticeInterview)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// rootCtx is cancelled on SIGINT/SIGTERM so background goroutines
	// (e.g. services.StartCleanup) can exit gracefully.
	rootCtx, rootCancel := context.WithCancel(context.Background())
	services.StartCleanup(rootCtx)
	services.StartDedupCleanup(rootCtx)

	go func() {
		<-quit
		log.Println("\n[Server] Shutting down gracefully...")
		rootCancel()
		// Shutdown timeout must exceed WriteTimeout (300s) so that in-flight
		// SSE streams (AI generations up to 5 minutes) are not forcibly cut
		// off during deployments/restarts — truncating user results and
		// wasting already-incurred AI API costs.
		ctx, cancel := context.WithTimeout(context.Background(), 310*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("[Server] Starting on port %s\n", port)
	if err := app.Listen(":" + port); err != nil {
		// R37b-B3: log.Fatalf calls os.Exit which skips all defers
		// (db.Close, verificationStore.Close, paymentHandler.Close).
		// Run cleanup manually before exiting so SQLite WAL checkpoint
		// executes and background goroutines are signalled to stop.
		log.Printf("Server failed to start: %v", err)
		rootCancel()
		db.Close()
		verificationStore.Close()
		paymentHandler.Close()
		middleware.FlushSentry()
		os.Exit(1)
	}
	// FlushSentry runs in the main goroutine after Listen returns (i.e.
	// after shutdown completes). The prior code ran FlushSentry inside
	// the signal-handler goroutine, but app.Listen could return and main
	// could exit before FlushSentry finished — losing buffered events.
	middleware.FlushSentry()
}

// R38b-L5: removed unused userStore field — all methods only use dp.db.
type DatabasePersistence struct {
	db *database.Database
}

func (dp *DatabasePersistence) Save(users map[string]*models.User) error {
	for _, user := range users {
		if err := dp.db.SaveUser(user); err != nil {
			return err
		}
	}
	return nil
}

// SaveUser persists a single user to SQLite without re-writing every other
// user. Use this when only one user changed (login, register, plan upgrade).
func (dp *DatabasePersistence) SaveUser(user *models.User) error {
	return dp.db.SaveUser(user)
}

// UpdateUserUsage performs a targeted UPDATE of only usage_count.
// Prefer this over SaveUser for persistUsage to avoid clobbering concurrent
// token/plan/max_free_usage updates with a stale snapshot.
func (dp *DatabasePersistence) UpdateUserUsage(email string, usageCount int) error {
	return dp.db.UpdateUserUsage(email, usageCount)
}

// AdjustUserUsage performs an incremental UPDATE (usage_count + delta).
// R52-B1: used by persistUsage to eliminate stale-write races.
func (dp *DatabasePersistence) AdjustUserUsage(email string, delta int) error {
	return dp.db.AdjustUserUsage(email, delta)
}

// UpdateUserToken performs a targeted UPDATE of only the token column.
// R37-B1: used by auth paths to avoid SaveUser clobbering usage_count.
func (dp *DatabasePersistence) UpdateUserToken(email, token string) error {
	return dp.db.UpdateUserToken(email, token)
}

// UpdateUserTokenAndPassword updates token plus (optionally) password and
// password_type in a single UPDATE. R37-B1: used by Login.
func (dp *DatabasePersistence) UpdateUserTokenAndPassword(email, token, password, passwordType string) error {
	return dp.db.UpdateUserTokenAndPassword(email, token, password, passwordType)
}

// UpdateUserPlan performs a targeted UPDATE of plan-related columns only.
// R37-B1: used by payment paths to avoid SaveUser clobbering usage_count.
func (dp *DatabasePersistence) UpdateUserPlan(email, plan, subscriptionID, captureID string, maxFreeUsage int) error {
	return dp.db.UpdateUserPlan(email, plan, subscriptionID, captureID, maxFreeUsage)
}

// UpdateUserTemplates performs a targeted UPDATE of only the
// purchased_templates and template_captures columns. R37-B1: used by
// CaptureTemplateOrder and the template-refund webhook to avoid clobbering
// concurrent usage_count/plan increments with a stale snapshot.
func (dp *DatabasePersistence) UpdateUserTemplates(email string, templates []string, captures []models.TemplateCapture) error {
	return dp.db.UpdateUserTemplates(email, templates, captures)
}
