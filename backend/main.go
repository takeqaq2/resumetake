package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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

const maxResumes = 5000

type Store struct {
	mu      sync.RWMutex
	resumes map[string]map[string]interface{}
	count   int64
}

var store = &Store{resumes: make(map[string]map[string]interface{})}
var startTime time.Time
var totalRequests int64

func (s *Store) Save(id string, data map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.resumes) >= maxResumes {
		oldest := ""
		for k := range s.resumes {
			oldest = k
			break
		}
		if oldest != "" {
			delete(s.resumes, oldest)
		}
	}
	s.resumes[id] = data
	atomic.AddInt64(&s.count, 1)
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
	return atomic.LoadInt64(&s.count)
}

type GroqRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
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

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
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
	"es": `Eres un consultor profesional de optimización de CV. Optimiza el CV basándote en la información del usuario y el puesto objetivo.

Requisitos:
1. Salida en español
2. Usa el método STAR (Situación-Tarea-Acción-Resultado)
3. Añade logros cuantificados
4. Extrae y coincide palabras clave ATS
5. Optimiza el resumen profesional

Formato JSON de retorno:
{
  "optimized_content": {
    "summary": "Resumen profesional optimizado",
    "experience": [{"company": "Empresa", "position": "Puesto", "duration": "Período", "highlights": ["Logro 1", "Logro 2"]}],
    "skills": ["Habilidad 1", "Habilidad 2"],
    "education": [{"school": "Universidad", "degree": "Título", "major": "Especialidad"}]
  },
  "ats_score": 85.0,
  "keywords": ["palabra1", "palabra2"],
  "suggestions": ["sugerencia1", "sugerencia2"]
}`,
	"pt": `Você é um consultor profissional de otimização de currículo. Otimize o currículo com base nas informações do usuário e na vaga-alvo.

Requisitos:
1. Saída em português
2. Use o método STAR (Situação-Tarefa-Ação-Resultado)
3. Adicione conquistas quantificadas
4. Extraia e combine palavras-chave ATS
5. Otimize o resumo profissional

Formato JSON de retorno:
{
  "optimized_content": {
    "summary": "Resumo profissional otimizado",
    "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Conquista 1", "Conquista 2"]}],
    "skills": ["Habilidade 1", "Habilidade 2"],
    "education": [{"school": "Universidade", "degree": "Diploma", "major": "Especialização"}]
  },
  "ats_score": 85.0,
  "keywords": ["palavra1", "palavra2"],
  "suggestions": ["sugestão1", "sugestão2"]
}`,
	"fr": `Vous êtes un consultant professionnel en optimisation de CV. Optimisez le CV en fonction des informations de l'utilisateur et du poste cible.

Exigences :
1. Sortie en français
2. Utilisez la méthode STAR (Situation-Tâche-Action-Résultat)
3. Ajoutez des réalisations quantifiées
4. Extrapolez et correspondez les mots-clés ATS
5. Optimisez le résumé professionnel

Format JSON de retour :
{
  "optimized_content": {
    "summary": "Résumé professionnel optimisé",
    "experience": [{"company": "Entreprise", "position": "Poste", "duration": "Période", "highlights": ["Réalisation 1", "Réalisation 2"]}],
    "skills": ["Compétence 1", "Compétence 2"],
    "education": [{"school": "Université", "degree": "Diplôme", "major": "Spécialité"}]
  },
  "ats_score": 85.0,
  "keywords": ["mot1", "mot2"],
  "suggestions": ["suggestion1", "suggestion2"]
}`,
	"de": `Sie sind ein professioneller Lebenslauf-Optimierungsberater. Optimieren Sie den Lebenslauf basierend auf den Informationen des Benutzers und der Zielposition.

Anforderungen:
1. Ausgabe auf Deutsch
2. Verwenden Sie die STAR-Methode (Situation-Task-Aktion-Ergebnis)
3. Fügen Sie quantifizierte Erfolge hinzu
4. Extrahieren und passen Sie ATS-Schlüsselwörter an
5. Optimieren Sie das Berufsprofil

JSON-Rückgabeformat:
{
  "optimized_content": {
    "summary": "Optimiertes Berufsprofil",
    "experience": [{"company": "Unternehmen", "position": "Position", "duration": "Zeitraum", "highlights": ["Erfolg 1", "Erfolg 2"]}],
    "skills": ["Fähigkeit 1", "Fähigkeit 2"],
    "education": [{"school": "Universität", "degree": "Abschluss", "major": "Fachgebiet"}]
  },
  "ats_score": 85.0,
  "keywords": ["Schlüsselwort1", "Schlüsselwort2"],
  "suggestions": ["Vorschlag1", "Vorschlag2"]
}`,
	"hi": `आप एक पेशेवर रिज़्यूमे ऑप्टिमाइज़ेशन सलाहकार हैं। उपयोगकर्ता द्वारा प्रदान की गई जानकारी और लक्षित पद के आधार पर रिज़्यूमे को ऑप्टिमाइज़ करें।

आवश्यकताएँ:
1. हिंदी में आउटपुट
2. STAR विधि (स्थिति-कार्य-कार्रवाई-परिणाम) का उपयोग करें
3. मात्रात्मक उपलब्धियाँ जोड़ें
4. ATS कीवर्ड निकालें और मैच करें
5. पेशेवर सारांश ऑप्टिमाइज़ करें

JSON प्रारूप में वापस करें:
{
  "optimized_content": {
    "summary": "ऑप्टिमाइज़ पेशेवर सारांश",
    "experience": [{"company": "कंपनी", "position": "पद", "duration": "अवधि", "highlights": ["उपलब्धि 1", "उपलब्धि 2"]}],
    "skills": ["कौशल 1", "कौशल 2"],
    "education": [{"school": "विश्वविद्यालय", "degree": "डिग्री", "major": "विषय"}]
  },
  "ats_score": 85.0,
  "keywords": ["कीवर्ड1", "कीवर्ड2"],
  "suggestions": ["सुझाव1", "सुझाव2"]
}`,
}

var perspectivePrompts = map[string]string{
	"zh": `你是一位专业的简历分析顾问。请根据用户提供的简历内容，从4个视角进行分析。

4个视角：
1. "original"（原始的我）：忠实展示简历当前的真实内容和水平，不做美化
2. "optimized"（优化后的我）：用STAR法则+量化成果+ATS关键词优化，展示最佳版本
3. "imagined"（我幻想的我）：大胆畅想如果用户有更多经验/技能，简历可以达到的理想状态
4. "desired"（HR希望的我）：站在HR角度，分析HR最想看到的内容和表达方式

请返回JSON格式：
{
  "original": {
    "summary": "当前简历的个人简介",
    "experience": [{"company": "公司", "position": "职位", "duration": "时间", "highlights": ["当前描述1", "描述2"]}],
    "skills": ["当前技能"],
    "score": 65.0,
    "analysis": "对当前简历的客观评价"
  },
  "optimized": {
    "summary": "优化后的个人简介",
    "experience": [{"company": "公司", "position": "职位", "duration": "时间", "highlights": ["优化后描述1", "描述2"]}],
    "skills": ["优化后技能"],
    "score": 85.0,
    "analysis": "优化策略说明"
  },
  "imagined": {
    "summary": "理想状态的个人简介",
    "experience": [{"company": "公司", "position": "职位", "duration": "时间", "highlights": ["理想描述1", "描述2"]}],
    "skills": ["理想技能"],
    "score": 95.0,
    "analysis": "如何达到理想状态的建议"
  },
  "desired": {
    "summary": "HR期望看到的个人简介",
    "experience": [{"company": "公司", "position": "职位", "duration": "时间", "highlights": ["HR想看到的描述1", "描述2"]}],
    "skills": ["HR关注的技能"],
    "score": 88.0,
    "analysis": "HR筛选简历的关注点"
  }
}`,
	"en": `You are a professional resume analyst. Analyze the user's resume from 4 perspectives.

4 perspectives:
1. "original": Faithfully show the current real content and level, no embellishment
2. "optimized": Use STAR method + quantified achievements + ATS keywords, show the best version
3. "imagined": Boldly imagine the ideal state if the user had more experience/skills
4. "desired": From HR's perspective, analyze what HR most wants to see

Return JSON format:
{
  "original": {
    "summary": "Current resume summary",
    "experience": [{"company": "Company", "position": "Title", "duration": "Period", "highlights": ["Current desc 1", "desc 2"]}],
    "skills": ["Current skills"],
    "score": 65.0,
    "analysis": "Objective evaluation of current resume"
  },
  "optimized": {
    "summary": "Optimized summary",
    "experience": [{"company": "Company", "position": "Title", "duration": "Period", "highlights": ["Optimized desc 1", "desc 2"]}],
    "skills": ["Optimized skills"],
    "score": 85.0,
    "analysis": "Optimization strategy explanation"
  },
  "imagined": {
    "summary": "Ideal state summary",
    "experience": [{"company": "Company", "position": "Title", "duration": "Period", "highlights": ["Ideal desc 1", "desc 2"]}],
    "skills": ["Ideal skills"],
    "score": 95.0,
    "analysis": "How to reach ideal state"
  },
  "desired": {
    "summary": "What HR wants to see",
    "experience": [{"company": "Company", "position": "Title", "duration": "Period", "highlights": ["HR wants to see 1", "desc 2"]}],
    "skills": ["HR-focused skills"],
    "score": 88.0,
    "analysis": "HR screening focus points"
  }
}`,
}

func getPrompt(lang string) string {
	if p, ok := prompts[lang]; ok {
		return p
	}
	return prompts["en"]
}

type AIProvider struct {
	Name    string
	BaseURL string
	Model   string
	APIKey  string
}

func getAIProviders() []AIProvider {
	var providers []AIProvider

	// SiliconFlow (China-accessible, 9B models free forever)
	if key := os.Getenv("SILICONFLOW_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "siliconflow",
			BaseURL: "https://api.siliconflow.cn/v1",
			Model:   "Qwen/Qwen2.5-7B-Instruct",
			APIKey:  key,
		})
	}

	// Zhipu GLM (China-accessible, GLM-4-Flash free forever)
	if key := os.Getenv("ZHIPU_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "zhipu",
			BaseURL: "https://open.bigmodel.cn/api/paas/v4",
			Model:   "glm-4-flash",
			APIKey:  key,
		})
	}

	// DeepSeek
	if key := os.Getenv("DEEPSEEK_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "deepseek",
			BaseURL: "https://api.deepseek.com",
			Model:   "deepseek-chat",
			APIKey:  key,
		})
	}

	// Doubao (Volcengine)
	if key := os.Getenv("DOUBAO_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "doubao",
			BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
			Model:   "doubao-1.5-pro-32k",
			APIKey:  key,
		})
	}

	// Groq
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "groq",
			BaseURL: "https://api.groq.com/openai/v1",
			Model:   "llama-3.3-70b-versatile",
			APIKey:  key,
		})
	}

	// Google Gemini (OpenAI-compatible endpoint, blocked in China)
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "gemini",
			BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
			Model:   "gemini-2.0-flash",
			APIKey:  key,
		})
	}

	// Cerebras
	if key := os.Getenv("CEREBRAS_API_KEY"); key != "" {
		providers = append(providers, AIProvider{
			Name:    "cerebras",
			BaseURL: "https://api.cerebras.ai/v1",
			Model:   "llama-3.3-70b",
			APIKey:  key,
		})
	}

	return providers
}

func callAIWithProvider(provider AIProvider, userMsg, lang string) (map[string]interface{}, error) {
	reqBody := GroqRequest{
		Model: provider.Model,
		Messages: []GroqMessage{
			{Role: "system", Content: getPrompt(lang)},
			{Role: "user", Content: userMsg},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("request preparation failed")
	}

	apiURL := provider.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s service unavailable", provider.Name)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response", provider.Name)
	}

	var groqResp GroqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return nil, fmt.Errorf("invalid %s response format", provider.Name)
	}
	if groqResp.Error != nil {
		return nil, fmt.Errorf("%s error: %s", provider.Name, groqResp.Error.Message)
	}
	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from %s", provider.Name)
	}

	content := groqResp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse %s result", provider.Name)
	}
	return result, nil
}

func callAI(resumeContent, targetJob, jobDescription, lang, moduleHints string) (map[string]interface{}, error) {
	userMsg := fmt.Sprintf("Target Position: %s\nJob Description: %s\nResume Content: %s", targetJob, jobDescription, resumeContent)
	if lang == "zh" {
		userMsg = fmt.Sprintf("目标职位: %s\n职位描述: %s\n简历内容: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "ja" {
		userMsg = fmt.Sprintf("希望職種: %s\n職務記述書: %s\n履歴書内容: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "ko" {
		userMsg = fmt.Sprintf("희망 직종: %s\n직무 설명: %s\n이력서 내용: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "ar" {
		userMsg = fmt.Sprintf("المنصب المستهدف: %s\nوصف الوظيفة: %s\nمحتوى السيرة الذاتية: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "es" {
		userMsg = fmt.Sprintf("Puesto objetivo: %s\nDescripción del puesto: %s\nContenido del CV: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "pt" {
		userMsg = fmt.Sprintf("Cargo alvo: %s\nDescrição da vaga: %s\nConteúdo do currículo: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "fr" {
		userMsg = fmt.Sprintf("Poste cible: %s\nDescription du poste: %s\nContenu du CV: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "de" {
		userMsg = fmt.Sprintf("Zielposition: %s\nStellenbeschreibung: %s\nLebenslauf-Inhalt: %s", targetJob, jobDescription, resumeContent)
	} else if lang == "hi" {
		userMsg = fmt.Sprintf("लक्षित पद: %s\nनौकरी विवरण: %s\nरिज़्यूमे सामग्री: %s", targetJob, jobDescription, resumeContent)
	}

	if moduleHints != "" {
		userMsg += "\n\nOptimization focus:\n" + moduleHints
	}

	providers := getAIProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no AI provider configured")
	}

	var lastErr error
	for _, p := range providers {
		result, err := callAIWithProvider(p, userMsg, lang)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("all AI providers failed: %s", lastErr.Error())
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	startTime = time.Now()

	app := fiber.New(fiber.Config{
		AppName:       "ResumeTake API v2.0",
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
		HSTSMaxAge:         31536000,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://resume.takee.top,http://localhost:5173",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowMethods: "GET,POST,OPTIONS",
		MaxAge:       86400,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		atomic.AddInt64(&totalRequests, 1)
		err := c.Next()
		latency := time.Since(start)
		c.Set("X-Process-Time", latency.String())
		if rid, ok := c.Locals("requestid").(string); ok {
			c.Set("X-Request-Id", rid)
		}
		return err
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		providers := getAIProviders()
		providerNames := make([]string, len(providers))
		for i, p := range providers {
			providerNames[i] = p.Name
		}
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
			"requests":  store.Count(),
			"total":     atomic.LoadInt64(&totalRequests),
			"version":   "2.0.0",
			"ai":        providerNames,
			"memory":    fmt.Sprintf("%d MB", getMemUsage()),
		})
	})

	v1 := app.Group("/api/v1")

	v1.Post("/resumes", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		title, _ := body["title"].(string)
		if strings.TrimSpace(title) == "" {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "title is required"})
		}
		if len(title) > 200 {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "title too long"})
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

	v1.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "NO_FILE", "message": "No file uploaded"})
		}
		if file.Size > 5*1024*1024 {
			return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "File too large (max 5MB)"})
		}
		ext := strings.ToLower(file.Filename)
		if !strings.HasSuffix(ext, ".txt") && !strings.HasSuffix(ext, ".pdf") && !strings.HasSuffix(ext, ".doc") && !strings.HasSuffix(ext, ".docx") {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_TYPE", "message": "Only .txt, .pdf, .doc, .docx files supported"})
		}
		f, err := file.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read file"})
		}
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read file content"})
		}
		text := string(content)
		if len(text) > 15000 {
			text = text[:15000]
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data": map[string]interface{}{
				"filename": file.Filename,
				"size":     file.Size,
				"text":     text,
			},
		})
	})

	v1.Post("/scrape-job", limiter.New(limiter.Config{
		Max:        20,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		jobURL, _ := body["url"].(string)
		if jobURL == "" {
			return c.Status(400).JSON(fiber.Map{"error": "NO_URL", "message": "URL is required"})
		}
		if !strings.HasPrefix(jobURL, "http://") && !strings.HasPrefix(jobURL, "https://") {
			jobURL = "https://" + jobURL
		}

		req, err := http.NewRequest("GET", jobURL, nil)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_URL", "message": "Invalid URL"})
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, err := httpClient.Do(req)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": "FETCH_FAILED", "message": "Failed to fetch job page"})
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return c.Status(502).JSON(fiber.Map{"error": "FETCH_FAILED", "message": fmt.Sprintf("HTTP %d", resp.StatusCode)})
		}

		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 500*1024))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read page"})
		}

		html := string(bodyBytes)
		title := extractMeta(html, "og:title")
		desc := extractMeta(html, "description")
		if desc == "" {
			desc = extractMeta(html, "og:description")
		}

		cleanText := stripHTML(html)
		if len(cleanText) > 5000 {
			cleanText = cleanText[:5000]
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": map[string]interface{}{
				"url":         jobURL,
				"title":       title,
				"description": desc,
				"text":        cleanText,
				"status":      resp.StatusCode,
			},
		})
	})

	v1.Post("/optimize", limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		lang, _ := body["lang"].(string)
		if lang == "" {
			lang = "en"
		}
		validLangs := map[string]bool{"zh": true, "en": true, "ja": true, "ko": true, "ar": true, "es": true, "pt": true, "fr": true, "de": true, "hi": true}
		if !validLangs[lang] {
			lang = "en"
		}
		targetJob, _ := body["target_job"].(string)
		jobDesc, _ := body["job_description"].(string)
		if len(targetJob) > 500 {
			targetJob = targetJob[:500]
		}
		if len(jobDesc) > 2000 {
			jobDesc = jobDesc[:2000]
		}

		var resumeContent string
		if rc, ok := body["resume_text"].(string); ok && rc != "" {
			resumeContent = rc
		} else if rc, _ := json.Marshal(body["resume_content"]); rc != nil && string(rc) != "null" {
			resumeContent = string(rc)
		}
		if len(resumeContent) > 10000 {
			resumeContent = resumeContent[:10000]
		}

		modules, _ := body["modules"].([]interface{})
		if len(modules) == 0 {
			modules = []interface{}{"ats", "star", "quant", "summary", "format"}
		}

		moduleHints := ""
		for _, m := range modules {
			switch m.(string) {
			case "ats":
				moduleHints += "1. Extract and match ATS keywords from the job description\n"
			case "star":
				moduleHints += "2. Rewrite work experience using STAR method (Situation-Task-Action-Result)\n"
			case "quant":
				moduleHints += "3. Add quantified achievements with specific metrics and data\n"
			case "summary":
				moduleHints += "4. Optimize professional summary to highlight core competencies\n"
			case "format":
				moduleHints += "5. Optimize resume structure, formatting, and layout\n"
			}
		}

		result, err := callAI(resumeContent, targetJob, jobDesc, lang, moduleHints)
		if err != nil {
			return c.Status(503).JSON(fiber.Map{
				"success": false,
				"error":   "Service temporarily unavailable, please try again later",
			})
		}

		return c.JSON(fiber.Map{"success": true, "data": result})
	})

	v1.Post("/perspective", limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		lang, _ := body["lang"].(string)
		if lang == "" {
			lang = "en"
		}
		validLangs := map[string]bool{"zh": true, "en": true, "ja": true, "ko": true, "ar": true, "es": true, "pt": true, "fr": true, "de": true, "hi": true}
		if !validLangs[lang] {
			lang = "en"
		}

		var resumeContent string
		if rc, ok := body["resume_text"].(string); ok && rc != "" {
			resumeContent = rc
		}
		if len(resumeContent) > 10000 {
			resumeContent = resumeContent[:10000]
		}
		if resumeContent == "" {
			return c.Status(400).JSON(fiber.Map{"error": "NO_CONTENT", "message": "Resume content is required"})
		}

		targetJob, _ := body["target_job"].(string)
		jobDesc, _ := body["job_description"].(string)
		if len(targetJob) > 500 {
			targetJob = targetJob[:500]
		}
		if len(jobDesc) > 2000 {
			jobDesc = jobDesc[:2000]
		}

		prompt, ok := perspectivePrompts[lang]
		if !ok {
			prompt = perspectivePrompts["en"]
		}

		userMsg := fmt.Sprintf("Target Position: %s\nJob Description: %s\nResume Content: %s", targetJob, jobDesc, resumeContent)
		if lang == "zh" {
			userMsg = fmt.Sprintf("目标职位: %s\n职位描述: %s\n简历内容: %s", targetJob, jobDesc, resumeContent)
		}

		providers := getAIProviders()
		if len(providers) == 0 {
			return c.Status(503).JSON(fiber.Map{"error": "NO_AI", "message": "No AI provider configured"})
		}

		var lastErr error
		for _, p := range providers {
			reqBody := GroqRequest{
				Model: p.Model,
				Messages: []GroqMessage{
					{Role: "system", Content: prompt},
					{Role: "user", Content: userMsg},
				},
				MaxTokens:   4096,
				Temperature: 0.7,
			}
			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				lastErr = err
				continue
			}
			apiURL := p.BaseURL + "/chat/completions"
			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				lastErr = err
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+p.APIKey)
			resp, err := httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("%s unavailable", p.Name)
				continue
			}
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = err
				continue
			}
			var groqResp GroqResponse
			if err := json.Unmarshal(bodyBytes, &groqResp); err != nil {
				lastErr = err
				continue
			}
			if groqResp.Error != nil {
				lastErr = fmt.Errorf("%s: %s", p.Name, groqResp.Error.Message)
				continue
			}
			if len(groqResp.Choices) == 0 {
				lastErr = fmt.Errorf("no response from %s", p.Name)
				continue
			}
			content := groqResp.Choices[0].Message.Content
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(content), &result); err != nil {
				lastErr = fmt.Errorf("failed to parse %s result", p.Name)
				continue
			}
			return c.JSON(fiber.Map{"success": true, "data": result})
		}
		return c.Status(503).JSON(fiber.Map{"success": false, "error": lastErr.Error()})
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
			"es": {
				{"id": "professional", "name": "Profesional", "description": "Para puestos tradicionales y empresariales"},
				{"id": "modern", "name": "Moderno", "description": "Para puestos tecnológicos"},
				{"id": "creative", "name": "Creativo", "description": "Para puestos de diseño y creativos"},
				{"id": "academic", "name": "Académico", "description": "Para puestos educativos e investigación"},
				{"id": "executive", "name": "Ejecutivo", "description": "Para puestos de alta dirección"},
				{"id": "minimal", "name": "Minimalista", "description": "Limpio y versátil"},
			},
			"pt": {
				{"id": "professional", "name": "Profissional", "description": "Para cargos tradicionais e empresariais"},
				{"id": "modern", "name": "Moderno", "description": "Para cargos de tecnologia"},
				{"id": "creative", "name": "Criativo", "description": "Para cargos de design e criativos"},
				{"id": "academic", "name": "Acadêmico", "description": "Para cargos educacionais e de pesquisa"},
				{"id": "executive", "name": "Executivo", "description": "Para cargos de alta diretoria"},
				{"id": "minimal", "name": "Minimalista", "description": "Limpo e versátil"},
			},
			"fr": {
				{"id": "professional", "name": "Professionnel", "description": "Pour les postes traditionnels et d'entreprise"},
				{"id": "modern", "name": "Moderne", "description": "Pour les postes technologiques"},
				{"id": "creative", "name": "Créatif", "description": "Pour les postes de design et créatifs"},
				{"id": "academic", "name": "Académique", "description": "Pour les postes éducatifs et de recherche"},
				{"id": "executive", "name": "Exécutif", "description": "Pour les postes de haute direction"},
				{"id": "minimal", "name": "Minimaliste", "description": "Épuré et polyvalent"},
			},
			"de": {
				{"id": "professional", "name": "Professionell", "description": "Für traditionelle und Geschäftspositionen"},
				{"id": "modern", "name": "Modern", "description": "Für Tech- und Startup-Positionen"},
				{"id": "creative", "name": "Kreativ", "description": "Für Design- und Kreativpositionen"},
				{"id": "academic", "name": "Akademisch", "description": "Für Bildungs- und Forschungspositionen"},
				{"id": "executive", "name": "Führungskraft", "description": "Für Senior-Management-Positionen"},
				{"id": "minimal", "name": "Minimalistisch", "description": "Aufgeräumt und vielseitig"},
			},
			"ar": {
				{"id": "professional", "name": "احترافي", "description": "للمناصب التقليدية والتجارية"},
				{"id": "modern", "name": "عصري", "description": "للمناصب التقنية"},
				{"id": "creative", "name": "إبداعي", "description": "لمناصب التصميم والإبداع"},
				{"id": "academic", "name": "أكاديمي", "description": "لمناصب التعليم والبحث"},
				{"id": "executive", "name": "تنفيذي", "description": "لمناصب الإدارة العليا"},
				{"id": "minimal", "name": "بسيط", "description": "أنيق وعملي"},
			},
			"hi": {
				{"id": "professional", "name": "पेशेवर", "description": "पारंपरिक और व्यावसायिक पदों के लिए"},
				{"id": "modern", "name": "आधुनिक", "description": "तकनीकी और स्टार्टअप पदों के लिए"},
				{"id": "creative", "name": "रचनात्मक", "description": "डिज़ाइन और रचनात्मक पदों के लिए"},
				{"id": "academic", "name": "शैक्षणिक", "description": "शिक्षा और अनुसंधान पदों के लिए"},
				{"id": "executive", "name": "कार्यकारी", "description": "वरिष्ठ प्रबंधन पदों के लिए"},
				{"id": "minimal", "name": "न्यूनतम", "description": "साफ और बहुमुखी"},
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

func getMemUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024
}

func extractMeta(html, name string) string {
	idx := strings.Index(html, name)
	if idx == -1 {
		return ""
	}
	rest := html[idx:]
	if eq := strings.Index(rest, "content=\""); eq != -1 {
		rest = rest[eq+9:]
		if end := strings.Index(rest, "\""); end != -1 {
			return rest[:end]
		}
	}
	return ""
}

func stripHTML(html string) string {
	var result strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			result.WriteString(" ")
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	text := result.String()
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", " ")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}
