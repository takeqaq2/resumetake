package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
)

type Store struct {
	mu      sync.RWMutex
	resumes map[string]map[string]interface{}
	count   int64
}

var store = &Store{resumes: make(map[string]map[string]interface{})}

func (s *Store) Save(id string, data map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resumes[id] = data
	s.count++
}

func (s *Store) Get(id string) (map[string]interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.resumes[id]
	return r, ok
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.resumes[id]; ok {
		delete(s.resumes, id)
		return true
	}
	return false
}

func (s *Store) Count() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	app := fiber.New(fiber.Config{
		AppName:       "ResumeTake API v1.0",
		BodyLimit:     10 * 1024 * 1024,
		ServerHeader:  "ResumeTake",
		StrictRouting: true,
		CaseSensitive: true,
		IdleTimeout:   30 * time.Second,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "${locals:requestid} ${method} ${path} ${status} - ${latency}\n",
	}))
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(helmet.New(helmet.Config{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		MaxAge:       86400,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        200,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		c.Set("X-Process-Time", latency.String())
		c.Set("X-Request-Id", c.Locals("requestid").(string))
		return err
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
			"requests":  store.Count(),
			"version":   "1.0.0",
		})
	})

	v1 := app.Group("/api/v1")

	v1.Post("/resumes", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "请求体格式错误"})
		}
		title, _ := body["title"].(string)
		if title == "" {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "title字段必填"})
		}
		id := uuid.New().String()
		now := time.Now().Format(time.RFC3339)
		data := map[string]interface{}{
			"id": id, "title": title, "content": body["content"],
			"created_at": now, "updated_at": now,
		}
		store.Save(id, data)
		return c.Status(201).JSON(fiber.Map{"success": true, "data": data})
	})

	v1.Get("/resumes/:id", func(c *fiber.Ctx) error {
		if r, ok := store.Get(c.Params("id")); ok {
			return c.JSON(fiber.Map{"success": true, "data": r})
		}
		return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "简历不存在"})
	})

	v1.Delete("/resumes/:id", func(c *fiber.Ctx) error {
		if store.Delete(c.Params("id")) {
			return c.JSON(fiber.Map{"success": true, "message": "删除成功"})
		}
		return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "简历不存在"})
	})

	v1.Post("/optimize", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "请求体格式错误"})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"optimized_content": fiber.Map{
					"summary": "资深专业人士，拥有丰富的行业经验和卓越的业绩记录。擅长团队协作与项目管理，具备出色的沟通能力和问题解决能力。",
					"experience": []fiber.Map{{
						"company": "示例科技有限公司", "position": "高级产品经理", "duration": "2020-至今",
						"highlights": []string{
							"领导5人团队完成核心产品开发，用户增长率达到150%",
							"优化产品流程，将交付周期缩短40%",
							"建立数据分析体系，驱动产品迭代决策",
						},
					}},
					"skills":    []string{"项目管理", "数据分析", "AI工具应用", "团队协作", "产品设计"},
					"education": []fiber.Map{{"school": "知名大学", "degree": "硕士学位", "major": "计算机科学与技术"}},
				},
				"ats_score":   87.3,
				"keywords":    []string{"项目管理", "数据分析", "AI应用", "团队协作", "产品优化", "流程改进"},
				"suggestions": []string{"建议添加更多量化成果", "突出AI工具使用经验", "个人简介加入核心关键词", "工作经历按STAR法则优化"},
			},
		})
	})

	v1.Get("/templates", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": []fiber.Map{
				{"id": "professional", "name": "专业商务", "description": "适合传统行业和商务岗位"},
				{"id": "modern", "name": "现代简约", "description": "适合互联网和科技行业"},
				{"id": "creative", "name": "创意设计", "description": "适合设计和创意岗位"},
				{"id": "academic", "name": "学术科研", "description": "适合教育和研究岗位"},
				{"id": "executive", "name": "高管专用", "description": "适合高级管理岗位"},
				{"id": "minimal", "name": "极简风格", "description": "简洁大方，通用性强"},
			},
		})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		_ = app.Shutdown()
	}()

	startTime = time.Now()
	_ = app.Listen(":" + port)
}

var startTime time.Time