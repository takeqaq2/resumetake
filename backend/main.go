package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
)

var resumesDB = make(map[string]map[string]interface{})
var requestCount int64

type Resume struct {
	ID               string                 `json:"id"`
	Title            string                 `json:"title"`
	Content          map[string]interface{} `json:"content"`
	TargetJob        string                 `json:"target_job,omitempty"`
	JobDescription   string                 `json:"job_description,omitempty"`
	OptimizedContent map[string]interface{} `json:"optimized_content,omitempty"`
	ATSScore         float64                `json:"ats_score,omitempty"`
	Keywords         []string               `json:"keywords,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

type OptimizeRequest struct {
	ResumeContent  map[string]interface{} `json:"resume_content"`
	TargetJob      string                 `json:"target_job,omitempty"`
	JobDescription string                 `json:"job_description,omitempty"`
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "ResumeTake API",
		BodyLimit:    10 * 1024 * 1024,
		ServerHeader: "ResumeTake",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		c.Set("X-Process-Time", time.Since(start).String())
		requestCount++
		return err
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":           "healthy",
			"timestamp":        time.Now().Format(time.RFC3339),
			"requests_served":  requestCount,
		})
	})

	v1 := app.Group("/api/v1")

	v1.Post("/resumes", func(c *fiber.Ctx) error {
		var resume Resume
		if err := c.BodyParser(&resume); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "参数解析失败"})
		}
		resume.ID = uuid.New().String()
		resume.CreatedAt = time.Now().Format(time.RFC3339)
		resume.UpdatedAt = resume.CreatedAt
		resumesDB[resume.ID] = map[string]interface{}{
			"id": resume.ID, "title": resume.Title, "content": resume.Content,
			"created_at": resume.CreatedAt, "updated_at": resume.UpdatedAt,
		}
		return c.Status(201).JSON(fiber.Map{"success": true, "data": resume})
	})

	v1.Get("/resumes/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if r, ok := resumesDB[id]; ok {
			return c.JSON(fiber.Map{"success": true, "data": r})
		}
		return c.Status(404).JSON(fiber.Map{"error": "简历不存在"})
	})

	v1.Delete("/resumes/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if _, ok := resumesDB[id]; ok {
			delete(resumesDB, id)
			return c.JSON(fiber.Map{"success": true, "message": "删除成功"})
		}
		return c.Status(404).JSON(fiber.Map{"error": "简历不存在"})
	})

	v1.Post("/optimize", func(c *fiber.Ctx) error {
		var req OptimizeRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "参数解析失败"})
		}

		result := fiber.Map{
			"success": true,
			"data": fiber.Map{
				"optimized_content": fiber.Map{
					"summary": "资深专业人士，拥有丰富的行业经验和卓越的业绩记录。擅长团队协作与项目管理，具备出色的沟通能力和问题解决能力。",
					"experience": []fiber.Map{{
						"company":   "示例科技有限公司",
						"position":  "高级产品经理",
						"duration":  "2020-至今",
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
				"suggestions": []string{"建议添加更多量化成果", "可以突出AI工具使用经验", "建议在个人简介中加入核心关键词", "工作经历建议按STAR法则优化"},
			},
		}
		return c.JSON(result)
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

	app.Listen(":8000")
}
