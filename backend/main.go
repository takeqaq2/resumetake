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
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		lang, _ := body["lang"].(string)
		if lang == "" {
			lang = "en"
		}
		responses := map[string]fiber.Map{
			"zh": {
				"optimized_content": fiber.Map{
					"summary": "资深专业人士，拥有丰富的行业经验和卓越的业绩记录。擅长团队协作与项目管理，具备出色的沟通能力和问题解决能力。",
					"experience": []fiber.Map{{"company": "示例科技有限公司", "position": "高级产品经理", "duration": "2020-至今", "highlights": []string{"领导5人团队完成核心产品开发，用户增长率达到150%", "优化产品流程，将交付周期缩短40%", "建立数据分析体系，驱动产品迭代决策"}}},
					"skills":    []string{"项目管理", "数据分析", "AI工具应用", "团队协作", "产品设计"},
					"education": []fiber.Map{{"school": "知名大学", "degree": "硕士学位", "major": "计算机科学与技术"}},
				},
				"ats_score": 87.3, "keywords": []string{"项目管理", "数据分析", "AI应用", "团队协作", "产品优化", "流程改进"},
				"suggestions": []string{"建议添加更多量化成果", "突出AI工具使用经验", "个人简介加入核心关键词", "工作经历按STAR法则优化"},
			},
			"en": {
				"optimized_content": fiber.Map{
					"summary": "Results-driven professional with extensive industry experience and a proven track record of excellence. Skilled in team collaboration, project management, and communication.",
					"experience": []fiber.Map{{"company": "Tech Corp", "position": "Senior Product Manager", "duration": "2020-Present", "highlights": []string{"Led 5-person team to launch core product, achieving 150% user growth", "Streamlined product workflow, reducing delivery cycle by 40%", "Built data analytics system to drive product iteration decisions"}}},
					"skills":    []string{"Project Management", "Data Analysis", "AI Tools", "Team Collaboration", "Product Design"},
					"education": []fiber.Map{{"school": "University of Technology", "degree": "Master's Degree", "major": "Computer Science"}},
				},
				"ats_score": 89.1, "keywords": []string{"Project Management", "Data Analysis", "AI Tools", "Team Leadership", "Product Optimization", "Process Improvement"},
				"suggestions": []string{"Add more quantified achievements", "Highlight AI tools experience", "Include core keywords in summary", "Optimize work experience using STAR method"},
			},
			"ja": {
				"optimized_content": fiber.Map{
					"summary": "豊富な業界経験と卓越した実績を持つプロフェッショナル。チームコラボレーション、プロジェクト管理、コミュニケーションに優れている。",
					"experience": []fiber.Map{{"company": "テック株式会社", "position": "シニアプロダクトマネージャー", "duration": "2020年〜現在", "highlights": []string{"5人のチームを率いてコア製品を開発、ユーザー成長率150%を達成", "製品ワークフローを最適化、納品サイクルを40%短縮", "データ分析システムを構築し、製品改善を推進"}}},
					"skills":    []string{"プロジェクト管理", "データ分析", "AIツール活用", "チームワーク", "プロダクト設計"},
					"education": []fiber.Map{{"school": "工科大学", "degree": "修士号", "major": "コンピュータサイエンス"}},
				},
				"ats_score": 86.5, "keywords": []string{"プロジェクト管理", "データ分析", "AI活用", "チームワーク", "プロダクト最適化", "プロセス改善"},
				"suggestions": []string{"量化された成果を追加", "AIツールの経験を強調", "概要にコアキーワードを含める", "STAR法で職務経歴を最適化"},
			},
			"ko": {
				"optimized_content": fiber.Map{
					"summary": "풍부한 산업 경험과 탁월한 실적을 보유한 결과 지향적 전문가. 팀 협업, 프로젝트 관리, 의사소통에 뛰어남.",
					"experience": []fiber.Map{{"company": "테크 주식회사", "position": "시니어 프로덕트 매니저", "duration": "2020-현재", "highlights": []string{"5명 팀을 이끌고 핵심 제품 출시, 사용자 성장률 150% 달성", "제품 워크플로우 최적화, 납기 주기 40% 단축", "데이터 분석 시스템 구축으로 제품 개선 추진"}}},
					"skills":    []string{"프로젝트 관리", "데이터 분석", "AI 도구 활용", "팀 협업", "제품 설계"},
					"education": []fiber.Map{{"school": "공과대학", "degree": "석사 학위", "major": "컴퓨터 과학"}},
				},
				"ats_score": 85.7, "keywords": []string{"프로젝트 관리", "데이터 분석", "AI 활용", "팀 리더십", "제품 최적화", "프로세스 개선"},
				"suggestions": []string{"정량화된 성과 추가", "AI 도구 경험 강조", "요약에 핵심 키워드 포함", "STAR 방법으로 직무 경험 최적화"},
			},
			"es": {
				"optimized_content": fiber.Map{
					"summary": "Profesional orientado a resultados con amplia experiencia industrial y un historial comprobado de excelencia.",
					"experience": []fiber.Map{{"company": "Tech Corp", "position": "Gerente de Producto Senior", "duration": "2020-Presente", "highlights": []string{"Lideró equipo de 5 personas para lanzar producto principal, logrando 150% de crecimiento", "Optimizó flujo de trabajo, reduciendo ciclo de entrega en 40%", "Construyó sistema de análisis de datos para impulsar decisiones"}}},
					"skills":    []string{"Gestión de Proyectos", "Análisis de Datos", "Herramientas IA", "Trabajo en Equipo", "Diseño de Producto"},
					"education": []fiber.Map{{"school": "Universidad de Tecnología", "degree": "Maestría", "major": "Ciencias de la Computación"}},
				},
				"ats_score": 84.2, "keywords": []string{"Gestión de Proyectos", "Análisis de Datos", "IA", "Liderazgo", "Optimización", "Mejora de Procesos"},
				"suggestions": []string{"Agregar logros cuantificados", "Destacar experiencia con IA", "Incluir palabras clave en resumen", "Optimizar experiencia con método STAR"},
			},
			"pt": {
				"optimized_content": fiber.Map{
					"summary": "Profissional orientado a resultados com ampla experiência industrial e histórico comprovado de excelência.",
					"experience": []fiber.Map{{"company": "Tech Corp", "position": "Gerente de Produto Sênior", "duration": "2020-Presente", "highlights": []string{"Liderou equipe de 5 pessoas no lançamento de produto principal, alcançando 150% de crescimento", "Otimizou fluxo de trabalho, reduzindo ciclo de entrega em 40%", "Construiu sistema de análise de dados para impulsionar decisões"}}},
					"skills":    []string{"Gestão de Projetos", "Análise de Dados", "Ferramentas IA", "Trabalho em Equipe", "Design de Produto"},
					"education": []fiber.Map{{"school": "Universidade de Tecnologia", "degree": "Mestrado", "major": "Ciência da Computação"}},
				},
				"ats_score": 83.8, "keywords": []string{"Gestão de Projetos", "Análise de Dados", "IA", "Liderança", "Otimização", "Melhoria de Processos"},
				"suggestions": []string{"Adicionar conquistas quantificadas", "Destacar experiência com IA", "Incluir palavras-chave no resumo", "Otimizar experiência com método STAR"},
			},
			"fr": {
				"optimized_content": fiber.Map{
					"summary": "Professionnel orienté résultats avec une vaste expérience industrielle et un bilan éprouvé d'excellence.",
					"experience": []fiber.Map{{"company": "Tech Corp", "position": "Chef de Produit Senior", "duration": "2020-Présent", "highlights": []string{"Dirigé une équipe de 5 personnes pour le lancement du produit phare, atteignant 150% de croissance", "Optimisé le flux de travail, réduisant le cycle de livraison de 40%", "Construit un système d'analyse de données pour guider les décisions"}}},
					"skills":    []string{"Gestion de Projet", "Analyse de Données", "Outils IA", "Travail d'Équipe", "Design de Produit"},
					"education": []fiber.Map{{"school": "Université de Technologie", "degree": "Master", "major": "Informatique"}},
				},
				"ats_score": 85.0, "keywords": []string{"Gestion de Projet", "Analyse de Données", "IA", "Leadership", "Optimisation", "Amélioration"},
				"suggestions": []string{"Ajouter des réalisations quantifiées", "Mettre en avant l'expérience IA", "Inclure mots-clés dans le résumé", "Optimiser l'expérience avec la méthode STAR"},
			},
			"de": {
				"optimized_content": fiber.Map{
					"summary": "Ergebnisorientierter Fachmann mit umfangreicher Branchenerfahrung und nachweislicher Erfolgsbilanz.",
					"experience": []fiber.Map{{"company": "Tech GmbH", "position": "Senior Product Manager", "duration": "2020-heute", "highlights": []string{"Führte 5-köpfiges Team bei Produktlaunch, 150% Nutzerwachstum erreicht", "Optimierte Produktworkflows, Lieferzyklus um 40% verkürzt", "Aufbaute Data Analytics System zur Steuerung von Produktentscheidungen"}}},
					"skills":    []string{"Projektmanagement", "Datenanalyse", "KI-Tools", "Teamarbeit", "Produktdesign"},
					"education": []fiber.Map{{"school": "Technische Universität", "degree": "Masterabschluss", "major": "Informatik"}},
				},
				"ats_score": 86.1, "keywords": []string{"Projektmanagement", "Datenanalyse", "KI", "Teamführung", "Produktoptimierung", "Prozessverbesserung"},
				"suggestions": []string{"Quantifizierte Erfolge hinzufügen", "KI-Erfahrung hervorheben", "Kernbegriffe im Profil einbeziehen", "Berufserfahrung mit STAR-Methode optimieren"},
			},
			"ar": {
				"optimized_content": fiber.Map{
					"summary": "محترف يركز على النتائج مع خبرة صناعية واسعة وسجل حافل بالتميز.",
					"experience": []fiber.Map{{"company": "شركة تك", "position": "مدير منتج أول", "duration": "2020-حتى الآن", "highlights": []string{"قاد فريق من 5 أشخاص لإطلاق المنتج الأساسي، محققاً نمو 150%", "حسّن سير العمل، قلّل دورة التسليم بنسبة 40%", "بنى نظام تحليل البيانات لتوجيه قرارات المنتج"}}},
					"skills":    []string{"إدارة المشاريع", "تحليل البيانات", "أدوات الذكاء الاصطناعي", "العمل الجماعي", "تصميم المنتج"},
					"education": []fiber.Map{{"school": "جامعة التكنولوجيا", "degree": "ماجستير", "major": "علوم الحاسوب"}},
				},
				"ats_score": 82.5, "keywords": []string{"إدارة المشاريع", "تحليل البيانات", "الذكاء الاصطناعي", "القيادة", "تحسين المنتج", "تحسين العمليات"},
				"suggestions": []string{"إضافة إنجازات مقاسة", "تسليط الضوء على خبرة الذكاء الاصطناعي", "تضمين الكلمات المفتاحية في الملخص", "تحسين الخبرة باستخدام طريقة STAR"},
			},
			"hi": {
				"optimized_content": fiber.Map{
					"summary": "परिणाम-उन्मुख पेशेवर जिसके पास व्यापक उद्योग अनुभव और उत्कृष्टता का सिद्ध ट्रैक रिकॉर्ड है।",
					"experience": []fiber.Map{{"company": "टेक कॉर्प", "position": "सीनियर प्रोडक्ट मैनेजर", "duration": "2020-वर्तमान", "highlights": []string{"5 सदस्यों की टीम का नेतृत्व कर मुख्य उत्पाद लॉन्च, 150% उपयोगकर्ता वृद्धि हासिल", "उत्पाद वर्कफ़्लो अनुकूलित, डिलीवरी चक्र 40% कम किया", "डेटा एनालिटिक्स सिस्टम बनाया जिससे उत्पाद निर्णयों को गति मिली"}}},
					"skills":    []string{"प्रोजेक्ट प्रबंधन", "डेटा विश्लेषण", "AI उपकरण", "टीम सहयोग", "उत्पाद डिज़ाइन"},
					"education": []fiber.Map{{"school": "प्रौद्योगिकी विश्वविद्यालय", "degree": "स्नातकोत्तर", "major": "कंप्यूटर विज्ञान"}},
				},
				"ats_score": 83.0, "keywords": []string{"प्रोजेक्ट प्रबंधन", "डेटा विश्लेषण", "AI", "नेतृत्व", "उत्पाद अनुकूलन", "प्रक्रिया सुधार"},
				"suggestions": []string{"मात्रात्मक उपलब्धियां जोड़ें", "AI अनुभव उजागर करें", "सारांश में मुख्य कीवर्ड शामिल करें", "STAR विधि से कार्य अनुभव अनुकूलित करें"},
			},
		}
		resp, ok := responses[lang]
		if !ok {
			resp = responses["en"]
		}
		return c.JSON(fiber.Map{"success": true, "data": resp})
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

	startTime = time.Now()
	_ = app.Listen(":" + port)
}

var startTime time.Time