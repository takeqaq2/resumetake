package main

import (
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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
		AppName:      "ResumeTake API",
		BodyLimit:    10 * 1024 * 1024,
		ServerHeader: "ResumeTake",
		StrictRouting: true,
		CaseSensitive: true,
	})

	app.Use(recover.New())
	app.Use(logger.New(":method :path :status - :res["content-type"] - :latency"))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		c.Set("X-Process-Time", time.Since(start).String())
		c.Set("X-Powered-By", "ResumeTake")
		return err
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"requests":  store.Count(),
			"version":   "1.0.0",
		})
	})

	v1 := app.Group("/api/v1")

	v1.Post("/resumes", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "参数解析失败"})
		}
		id := uuid.New().String()
		now := time.Now().Format(time.RFC3339)
		data := map[string]interface{}{
			"id": id, "title": body["title"], "content": body["content"],
			"created_at": now, "updated_at": now,
		}
		store.Save(id, data)
		return c.Status(201).JSON(fiber.Map{"success": true, "data": data})
	})

	v1.Get("/resumes/:id", func(c *fiber.Ctx) error {
		if r, ok := store.Get(c.Params("id")); ok {
			return c.JSON(fiber.Map{"success": true, "data": r})
		}
		return c.Status(404).JSON(fiber.Map{"error": "简历不存在"})
	})

	v1.Delete("/resumes/:id", func(c *fiber.Ctx) error {
		if store.Delete(c.Params("id")) {
			return c.JSON(fiber.Map{"success": true})
		}
		return c.Status(404).JSON(fiber.Map{"error": "简历不存在"})
	})

	v1.Post("/optimize", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "参数解析失败"})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"optimized_content": fiber.Map{
					"summary": "资深专业人士，拥有丰富的行业经验和卓越的业绩记录。擅长团队协作与项目管理。",
					"experience": []fiber.Map{{
						"company": "示例科技有限公司", "position": "高级产品经理", "duration": "2020-至今",
						"highlights": []string{"领导5人团队完成核心产品开发，用户增长150%", "优化产品流程，交付周期缩短40%", "建立数据分析体系，驱动产品迭代"},
					}},
					"skills": []string{"项目管理", "数据分析", "AI工具应用", "团队协作", "产品设计"},
				},
				"ats_score":   87.3,
				"keywords":    []string{"项目管理", "数据分析", "AI应用", "团队协作", "产品优化"},
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

	app.Listen(":" + port)
}
