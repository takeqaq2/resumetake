package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
var startTime time.Time

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

type GroqRequest struct {
	Model    string          `json:"model"`
	Messages []GroqMessage   `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
	Temperature float64     `json:"temperature,omitempty"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

var prompts = map[string]string{
	"zh": `你是一位专业的简历优化顾问。请根据用户提供的简历信息和目标职位，优化简历内容。

要求：
1. 用中文输出
2. 按照STAR法则（情境-任务-行动-结果）优化工作经历描述
3. 添加量化成果和数据
4. 提取并匹配ATS关键词
5. 优化个人简介，突出核心竞争力

请返回JSON格式：
{
  "optimized_content": {
    "summary": "优化后的个人简介",
    "experience": [{"company": "公司名", "position": "职位", "duration": "时间段", "highlights": ["优化后的成就描述1", "描述2"]}],
    "skills": ["技能1", "技能2"],
    "education": [{"school": "学校", "degree": "学位", "major": "专业"}]
  },
  "ats_score": 85.0,
  "keywords": ["关键词1", "关键词2"],
  "suggestions": ["建议1", "建议2"]
}`,
	"en": `You are a professional resume optimization consultant. Optimize the resume based on the user's information and target job.

Requirements:
1. Output in English
2. Use STAR method (Situation-Task-Action-Result) for work experience
3. Add quantified achievements with metrics
4. Extract and match ATS keywords
5. Optimize professional summary

Return JSON format:
{
  "optimized_content": {
    "summary": "Optimized professional summary",
    "experience": [{"company": "Company", "position": "Title", "duration": "Period", "highlights": ["Achievement 1", "Achievement 2"]}],
    "skills": ["Skill 1", "Skill 2"],
    "education": [{"school": "University", "degree": "Degree", "major": "Major"}]
  },
  "ats_score": 85.0,
  "keywords": ["keyword1", "keyword2"],
  "suggestions": ["suggestion1", "suggestion2"]
}`,
	"ja": `あなたはプロの履歴書最適化コンサルタントです。ユーザーの履歴書情報と希望職種に基づいて、履歴書を最適化してください。

要件：
1. 日本語で出力
2. STAR法（状況-課題-行動-結果）で職務経歴を最適化
3. 数値成果を追加
4. ATSキーワードを抽出・マッチング
5. 自己PRを最適化

JSON形式で返してください：
{
  "optimized_content": {
    "summary": "最適化された自己PR",
    "experience": [{"company": "会社名", "position": "役職", "duration": "期間", "highlights": ["成果1", "成果2"]}],
    "skills": ["スキル1", "スキル2"],
    "education": [{"school": "大学", "degree": "学位", "major": "専攻"}]
  },
  "ats_score": 85.0,
  "keywords": ["キーワード1", "キーワード2"],
  "suggestions": ["提案1", "提案2"]
}`,
	"ko": `당신은 전문 이력서 최적화 컨설턴트입니다. 사용자의 이력서 정보와 희망 직종을 기반으로 이력서를 최적화해주세요.

요구사항:
1. 한국어로 출력
2. STAR 방법(상황-과제-행동-결과)으로 업무 경험 최적화
3. 정량화된 성과 추가
4. ATS 키워드 추출 및 매칭
5. 자기소개 최적화

JSON 형식으로 반환:
{
  "optimized_content": {
    "summary": "최적화된 자기소개",
    "experience": [{"company": "회사명", "position": "직책", "duration": "기간", "highlights": ["성과1", "성과2"]}],
    "skills": ["기술1", "기술2"],
    "education": [{"school": "대학", "degree": "학위", "major": "전공"}]
  },
  "ats_score": 85.0,
  "keywords": ["키워드1", "키워드2"],
  "suggestions": ["제안1", "제안2"]
}`,
	"ar": `أنت مستشار متخصص في تحسين السيرة الذاتية. قم بتحسين السيرة الذاتية بناءً على معلومات المستخدم والمنصب المستهدف.

المتطلبات:
1. باللغة العربية
2. استخدم طريقة STAR للمهام العملية
3. أضف إنجازات مقاسة
4. استخرج كلمات ATS المفتاحية
5. حسّن الملخص المهني

أرجع بالتنسيق JSON:
{
  "optimized_content": {
    "summary": "ملخص مهني محسّن",
    "experience": [{"company": "الشركة", "position": "المنصب", "duration": "الفترة", "highlights": ["إنجاز1", "إنجاز2"]}],
    "skills": ["مهارة1", "مهارة2"],
    "education": [{"school": "الجامعة", "degree": "الدرجة", "major": "التخصص"}]
  },
  "ats_score": 85.0,
  "keywords": ["كلمة1", "كلمة2"],
  "suggestions": ["اقتراح1", "اقتراح2"]
}`,
}

func getPrompt(lang string) string {
	if p, ok := prompts[lang]; ok {
		return p
	}
	return prompts["en"]
}

func callGroqAI(resumeContent, targetJob, jobDescription, lang string) (map[string]interface{}, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY not set")
	}

	userMsg := fmt.Sprintf("目标职位: %s\n职位描述: %s\n简历内容: %s", targetJob, jobDescription, resumeContent)
	if lang == "en" {
		userMsg = fmt.Sprintf("Target Position: %s\nJob Description: %s\nResume Content: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "ja" {
		userMsg = fmt.Sprintf("希望職種: %s\n職務記述書: %s\n履歴書内容: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "ko" {
		userMsg = fmt.Sprintf("희망 직종: %s\n직무 설명: %s\n이력서 내용: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "ar" {
		userMsg = fmt.Sprintf("المنصب المستهدف: %s\nوصف الوظيفة: %s\nمحتوى السيرة الذاتية: %s", targetJob, jobDescription, resumeContent)
	}

	reqBody := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []GroqMessage{
			{Role: "system", Content: getPrompt(lang)},
			{Role: "user", Content: userMsg},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return nil, err
	}
	if groqResp.Error != nil {
		return nil, fmt.Errorf("groq error: %s", groqResp.Error.Message)
	}
	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	content := groqResp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}
	return result, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	startTime = time.Now()

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
		Format: "${locals:requestid} ${method} ${path} ${status} - ${latency}",
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
			"version":   "1.1.0",
			"ai":        "groq-free",
		})
	})

	v1 := app.Group("/api/v1")

	v1.Post("/resumes", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		title, _ := body["title"].(string)
		if title == "" {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "title is required"})
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
		return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "Resume not found"})
	})

	v1.Delete("/resumes/:id", func(c *fiber.Ctx) error {
		if store.Delete(c.Params("id")) {
			return c.JSON(fiber.Map{"success": true, "message": "Deleted"})
		}
		return c.Status(404).JSON(fiber.Map{"error": "NOT_FOUND", "message": "Resume not found"})
	})

	v1.Post("/optimize", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		lang, _ := body["lang"].(string)
		if lang == "" {
			lang = "en"
		}
		targetJob, _ := body["target_job"].(string)
		jobDesc, _ := body["job_description"].(string)
		resumeContent, _ := json.Marshal(body["resume_content"])
		if resumeContent == nil {
			resumeContent = []byte("{}")
		}

		store.mu.Lock()
		store.count++
		store.mu.Unlock()

		result, err := callGroqAI(string(resumeContent), targetJob, jobDesc, lang)
		if err != nil {
			return c.JSON(fiber.Map{
				"success": false,
				"error":   "AI optimization failed: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{"success": true, "data": result})
	})

	v1.Get("/templates", func(c *fiber.Ctx) error {
		lang := c.Query("lang", "en")
		templateData := map[string][]fiber.Map{
			"zh": {
				{"id": "professional", "name": "专业商务", "description": "适合传统行业和商务岗位"},
				{"id": "modern", "name": "现代简约", "description": "适合互联网和科技行业"},
				{"id": "creative", "name": "创意设计", "description": "适合设计和创意岗位"},
				{"id": "academic", "name": "学术科研", "description": "适合教育和研究岗位"},
				{"id": "executive", "name": "高管专用", "description": "适合高级管理岗位"},
				{"id": "minimal", "name": "极简风格", "description": "简洁大方，通用性强"},
			},
			"en": {
				{"id": "professional", "name": "Professional", "description": "For traditional and business roles"},
				{"id": "modern", "name": "Modern", "description": "For tech and startup roles"},
				{"id": "creative", "name": "Creative", "description": "For design and creative roles"},
				{"id": "academic", "name": "Academic", "description": "For education and research roles"},
				{"id": "executive", "name": "Executive", "description": "For senior management roles"},
				{"id": "minimal", "name": "Minimal", "description": "Clean and versatile"},
			},
			"ja": {
				{"id": "professional", "name": "プロフェッショナル", "description": "伝統的・ビジネス職向け"},
				{"id": "modern", "name": "モダン", "description": "IT・テック業界向け"},
				{"id": "creative", "name": "クリエイティブ", "description": "デザイン・クリエイティブ職向け"},
				{"id": "academic", "name": "アカデミック", "description": "教育・研究職向け"},
				{"id": "executive", "name": "エグゼクティブ", "description": "上級管理職向け"},
				{"id": "minimal", "name": "ミニマル", "description": "シンプルで汎用性が高い"},
			},
			"ko": {
				{"id": "professional", "name": "프로페셔널", "description": "전통적 및 비즈니스 직무용"},
				{"id": "modern", "name": "모던", "description": "IT 및 테크 업계용"},
				{"id": "creative", "name": "크리에이티브", "description": "디자인 및 크리에이티브 직무용"},
				{"id": "academic", "name": "아카데믹", "description": "교육 및 연구 직무용"},
				{"id": "executive", "name": "임원급", "description": "고위 경영진용"},
				{"id": "minimal", "name": "미니멀", "description": "깔끔하고 범용적"},
			},
		}
		data, ok := templateData[lang]
		if !ok {
			data = templateData["en"]
		}
		return c.JSON(fiber.Map{"success": true, "data": data})
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		_ = app.Shutdown()
	}()

	_ = app.Listen(":" + port)
}
