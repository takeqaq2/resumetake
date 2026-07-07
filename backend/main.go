package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"

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
const maxUploadBytes = 1 * 1024 * 1024

var userDataPath string

var allowedUploadExt = map[string]bool{
	".txt": true,
	".md":  true,
}

type Store struct {
	mu      sync.RWMutex
	resumes map[string]map[string]interface{}
	count   int64
}

var store = &Store{resumes: make(map[string]map[string]interface{})}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Name         string    `json:"name"`
	Token        string    `json:"token"`
	UsageCount   int       `json:"usage_count"`
	MaxFreeUsage int       `json:"max_free_usage"`
	CreatedAt    time.Time `json:"created_at"`
}

type persistedUser struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	Name         string    `json:"name"`
	Token        string    `json:"token"`
	UsageCount   int       `json:"usage_count"`
	MaxFreeUsage int       `json:"max_free_usage"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserStore struct {
	mu    sync.RWMutex
	users map[string]*User // keyed by email
}

var userStore = &UserStore{users: make(map[string]*User)}

func getUserDataPath() string {
	if path := os.Getenv("USER_DATA_FILE"); path != "" {
		return path
	}
	return "/app/data/users.json"
}

func (us *UserStore) GetByEmail(email string) (*User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()
	u, ok := us.users[email]
	return u, ok
}

func (us *UserStore) GetByToken(token string) (*User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()
	for _, u := range us.users {
		if u.Token == token {
			return u, true
		}
	}
	return nil, false
}

func (us *UserStore) Count() int {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return len(us.users)
}

func (us *UserStore) Save(user *User) {
	us.mu.Lock()
	us.users[user.Email] = user
	us.mu.Unlock()
	persistUsers()
}

func (us *UserStore) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	loaded := make(map[string]*persistedUser)
	if err := json.Unmarshal(b, &loaded); err != nil {
		return err
	}
	users := make(map[string]*User, len(loaded))
	for email, user := range loaded {
		users[email] = &User{
			ID:           user.ID,
			Email:        user.Email,
			Password:     user.Password,
			Name:         user.Name,
			Token:        user.Token,
			UsageCount:   user.UsageCount,
			MaxFreeUsage: user.MaxFreeUsage,
			CreatedAt:    user.CreatedAt,
		}
	}
	us.mu.Lock()
	us.users = users
	if us.users == nil {
		us.users = make(map[string]*User)
	}
	us.mu.Unlock()
	return nil
}

func (us *UserStore) Persist(path string) error {
	us.mu.RLock()
	snapshot := make(map[string]*persistedUser, len(us.users))
	for email, user := range us.users {
		snapshot[email] = &persistedUser{
			ID:           user.ID,
			Email:        user.Email,
			Password:     user.Password,
			Name:         user.Name,
			Token:        user.Token,
			UsageCount:   user.UsageCount,
			MaxFreeUsage: user.MaxFreeUsage,
			CreatedAt:    user.CreatedAt,
		}
	}
	us.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func persistUsers() {
	if userDataPath == "" {
		return
	}
	if err := userStore.Persist(userDataPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to persist users: %v\n", err)
	}
}

func generateToken(email string) string {
	h := sha256.Sum256([]byte(email + time.Now().String() + uuid.New().String()))
	return hex.EncodeToString(h[:])
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Missing authorization header"})
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Invalid authorization format"})
	}
	user, ok := userStore.GetByToken(token)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "UNAUTHORIZED", "message": "Invalid token"})
	}
	c.Locals("user", user)
	return c.Next()
}

type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	challenge := strings.TrimRight(strings.ToLower(string(fromServer)), ": ")
	switch challenge {
	case "username":
		return []byte(a.username), nil
	case "password":
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
}

func loginAuthFunc(username, password string) smtp.Auth {
	return &loginAuth{username: username, password: password}
}

type VerificationCode struct {
	Code      string
	CreatedAt time.Time
}

type VerificationStore struct {
	mu    sync.RWMutex
	codes map[string]*VerificationCode // keyed by email
}

var verificationStore = &VerificationStore{codes: make(map[string]*VerificationCode)}

func (vs *VerificationStore) Save(email, code string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.codes[email] = &VerificationCode{Code: code, CreatedAt: time.Now()}
}

func (vs *VerificationStore) Verify(email, code string) bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vc, ok := vs.codes[email]
	if !ok {
		return false
	}
	if time.Since(vc.CreatedAt) > 5*time.Minute {
		delete(vs.codes, email)
		return false
	}
	if vc.Code != code {
		return false
	}
	delete(vs.codes, email)
	return true
}

func generateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func sendVerificationEmail(toEmail, code string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		fmt.Printf("[SMTP] SMTP not configured, skipping email to %s, code: %s\n", toEmail, code)
		return nil
	}
	if smtpFrom == "" {
		smtpFrom = smtpUser
	}

	subject := "Subject: ResumeTake Verification Code\r\n"
	contentType := "MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n"
	body := fmt.Sprintf(`<html><body><div style="font-family:Arial,sans-serif;max-width:400px;margin:0 auto;padding:20px">
<h2 style="color:#4F46E5">ResumeTake</h2>
<p>Your verification code is:</p>
<div style="font-size:32px;font-weight:bold;color:#4F46E5;letter-spacing:8px;text-align:center;padding:20px;background:#F3F4F6;border-radius:8px">%s</div>
<p style="color:#6B7280;font-size:14px">This code expires in 5 minutes.</p>
</div></body></html>`, code)

	msg := []byte(subject + contentType + body)

	addr := smtpHost + ":" + smtpPort
	auth := loginAuthFunc(smtpUser, smtpPass)

	tlsconfig := &tls.Config{ServerName: smtpHost}
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = client.Mail(smtpFrom); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

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
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
		ForceAttemptHTTP2:   false,
	},
}

var prompts = map[string]string{
	"zh": `重要：你必须只输出一个有效的JSON对象。JSON前后不要有任何文字、解释或markdown。只输出纯JSON。

你是一位专业的简历优化顾问。请根据用户提供的简历信息和目标职位，优化简历内容。

要求：
1. 用中文输出
2. 按照STAR法则（情境-任务-行动-结果）优化工作经历描述
3. 添加量化成果和数据
4. 提取并匹配ATS关键词
5. 优化个人简介，突出核心竞争力

请只返回以下JSON结构：
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
	"en": `CRITICAL: You MUST respond with ONLY a valid JSON object. No text before or after the JSON. No markdown. No explanation.

You are a professional resume optimization consultant. Optimize the resume based on the user's information and target job.

Requirements:
1. Output in English
2. Use STAR method (Situation-Task-Action-Result) for work experience
3. Add quantified achievements with metrics
4. Extract and match ATS keywords
5. Optimize professional summary

Return ONLY this JSON structure:
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
	"ja": `あなたはプロの履歴書アナリストです。ユーザーの履歴書を4つの視点から分析してください。

4つの視点：
1. "original"（今の私）：履歴書の現在の内容とレベルを正直に表現
2. "optimized"（最適化された私）：STAR法+具体的な成果+ATSキーワードで最適化
3. "imagined"（理想の私）：もしもっと経験やスキルがあれば、到達できる理想の状態
4. "desired"（HRが求める私）：HR担当者の視点から、最も見てほしい内容

JSON形式で返してください：
{
  "original": {"summary": "現在の自己紹介", "experience": [{"company": "会社名", "position": "役職", "duration": "期間", "highlights": ["現在の説明1", "説明2"]}], "skills": ["現在のスキル"], "score": 65.0, "analysis": "現在の履歴書の客観的評価"},
  "optimized": {"summary": "最適化された自己紹介", "experience": [{"company": "会社名", "position": "役職", "duration": "期間", "highlights": ["最適化された説明1", "説明2"]}], "skills": ["最適化されたスキル"], "score": 85.0, "analysis": "最適化戦略の説明"},
  "imagined": {"summary": "理想の自己紹介", "experience": [{"company": "会社名", "position": "役職", "duration": "期間", "highlights": ["理想の説明1", "説明2"]}], "skills": ["理想のスキル"], "score": 95.0, "analysis": "理想の状態に到達する方法"},
  "desired": {"summary": "HRが見てほしい内容", "experience": [{"company": "会社名", "position": "役職", "duration": "期間", "highlights": ["HRが見てほしい説明1", "説明2"]}], "skills": ["HRが重視するスキル"], "score": 88.0, "analysis": "HRの履歴書チェックポイント"}
}`,
	"ko": `당신은 전문 이력서 분석가입니다. 사용자의 이력서를 4가지 관점에서 분석해주세요.

4가지 관점：
1. "original"（지금의 나）：이력서의 현재 내용과 수준을 솔직하게 표현
2. "optimized"（최적화된 나）：STAR 방법+정량화된 성과+ATS 키워드로 최적화
3. "imagined"（이상적인 나）：더 많은 경험과 기술이 있다면 도달할 수 있는 이상적인 상태
4. "desired"（HR이 원하는 나）：HR 관점에서 가장 보고 싶어하는 내용

JSON 형식으로 반환해주세요：
{
  "original": {"summary": "현재 자기소개", "experience": [{"company": "회사명", "position": "직책", "duration": "기간", "highlights": ["현재 설명1", "설명2"]}], "skills": ["현재 기술"], "score": 65.0, "analysis": "현재 이력서의 객관적 평가"},
  "optimized": {"summary": "최적화된 자기소개", "experience": [{"company": "회사명", "position": "직책", "duration": "기간", "highlights": ["최적화된 설명1", "설명2"]}], "skills": ["최적화된 기술"], "score": 85.0, "analysis": "최적화 전략 설명"},
  "imagined": {"summary": "이상적인 자기소개", "experience": [{"company": "회사명", "position": "직책", "duration": "기간", "highlights": ["이상적인 설명1", "설명2"]}], "skills": ["이상적인 기술"], "score": 95.0, "analysis": "이상적인 상태에 도달하는 방법"},
  "desired": {"summary": "HR이 보고 싶어하는 내용", "experience": [{"company": "회사명", "position": "직책", "duration": "기간", "highlights": ["HR이 보고 싶어하는 설명1", "설명2"]}], "skills": ["HR이 중시하는 기술"], "score": 88.0, "analysis": "HR 이력서 검토 포인트"}
}`,
	"es": `Eres un analista profesional de currículums. Analiza el currículum del usuario desde 4 perspectivas.

4 perspectivas：
1. "original"（yo original）：Muestra fielmente el contenido y nivel actual sin adornos
2. "optimized"（yo optimizado）：Usa método STAR + logros cuantificados + palabras clave ATS
3. "imagined"（yo imaginado）：Imagina el estado ideal si tuviera más experiencia/habilidades
4. "desired"（lo que HR quiere ver）：Desde la perspectiva de RRHH, analiza qué quieren ver

Devuelve en formato JSON：
{
  "original": {"summary": "Resumen actual del currículum", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Descripción actual 1", "Descripción 2"]}], "skills": ["Habilidades actuales"], "score": 65.0, "analysis": "Evaluación objetiva del currículum actual"},
  "optimized": {"summary": "Resumen optimizado", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Descripción optimizada 1", "Descripción 2"]}], "skills": ["Habilidades optimizadas"], "score": 85.0, "analysis": "Explicación de la estrategia de optimización"},
  "imagined": {"summary": "Resumen del estado ideal", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Descripción ideal 1", "Descripción 2"]}], "skills": ["Habilidades ideales"], "score": 95.0, "analysis": "Cómo alcanzar el estado ideal"},
  "desired": {"summary": "Lo que RRHH quiere ver", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Lo que RRHH quiere ver 1", "Descripción 2"]}], "skills": ["Habilidades que valora RRHH"], "score": 88.0, "analysis": "Puntos clave de revisión de RRHH"}
}`,
	"pt": `Você é um analista profissional de currículos. Analise o currículo do usuário a partir de 4 perspectivas.

4 perspectivas：
1. "original"（eu original）：Mostre fielmente o conteúdo e nível atual sem embelezamento
2. "optimized"（eu otimizado）：Use método STAR + conquistas quantificadas + palavras-chave ATS
3. "imagined"（eu imaginado）：Imagine o estado ideal se tivesse mais experiência/habilidades
4. "desired"（o que RH quer ver）：Da perspectiva de RH, analise o que mais querem ver

Retorne no formato JSON：
{
  "original": {"summary": "Resumo atual do currículo", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Descrição atual 1", "Descrição 2"]}], "skills": ["Habilidades atuais"], "score": 65.0, "analysis": "Avaliação objetiva do currículo atual"},
  "optimized": {"summary": "Resumo otimizado", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Descrição otimizada 1", "Descrição 2"]}], "skills": ["Habilidades otimizadas"], "score": 85.0, "analysis": "Explicação da estratégia de otimização"},
  "imagined": {"summary": "Resumo do estado ideal", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["Descrição ideal 1", "Descrição 2"]}], "skills": ["Habilidades ideais"], "score": 95.0, "analysis": "Como alcançar o estado ideal"},
  "desired": {"summary": "O que RH quer ver", "experience": [{"company": "Empresa", "position": "Cargo", "duration": "Período", "highlights": ["O que RH quer ver 1", "Descrição 2"]}], "skills": ["Habilidades que RH valoriza"], "score": 88.0, "analysis": "Pontos-chave de revisão de RH"}
}`,
	"fr": `Vous êtes un analyste professionnel de CV. Analysez le CV de l'utilisateur sous 4 perspectives.

4 perspectives：
1. "original"（moi original）：Montrez fidèlement le contenu et le niveau actuel sans embellissement
2. "optimized"（moi optimisé）：Utilisez la méthode STAR + réalisations quantifiées + mots-clés ATS
3. "imagined"（moi imaginé）：Imaginez l'état idéal avec plus d'expérience/compétences
4. "desired"（ce que RH veut voir）：Du point de vue RH, analysez ce qu'ils veulent voir

Retournez au format JSON：
{
  "original": {"summary": "Résumé actuel du CV", "experience": [{"company": "Entreprise", "position": "Poste", "duration": "Période", "highlights": ["Description actuelle 1", "Description 2"]}], "skills": ["Compétences actuelles"], "score": 65.0, "analysis": "Évaluation objective du CV actuel"},
  "optimized": {"summary": "Résumé optimisé", "experience": [{"company": "Entreprise", "position": "Poste", "duration": "Période", "highlights": ["Description optimisée 1", "Description 2"]}], "skills": ["Compétences optimisées"], "score": 85.0, "analysis": "Explication de la stratégie d'optimisation"},
  "imagined": {"summary": "Résumé de l'état idéal", "experience": [{"company": "Entreprise", "position": "Poste", "duration": "Période", "highlights": ["Description idéale 1", "Description 2"]}], "skills": ["Compétences idéales"], "score": 95.0, "analysis": "Comment atteindre l'état idéal"},
  "desired": {"summary": "Ce que RH veut voir", "experience": [{"company": "Entreprise", "position": "Poste", "duration": "Période", "highlights": ["Ce que RH veut voir 1", "Description 2"]}], "skills": ["Compétences valorisées par RH"], "score": 88.0, "analysis": "Points clés de vérification RH"}
}`,
	"de": `Sie sind ein professioneller Lebenslauf-Analyst. Analysieren Sie den Lebenslauf des Benutzers aus 4 Perspektiven.

4 Perspektiven：
1. "original"（das Original-Ich）：Zeigen Sie den aktuellen Inhalt und das aktuelle Niveau ohne Ausschmückung
2. "optimized"（das optimierte Ich）：Verwenden Sie STAR-Methode + quantifizierte Ergebnisse + ATS-Schlüsselwörter
3. "imagined"（das vorgestellte Ich）：Stellen Sie sich den idealen Zustand vor, wenn mehr Erfahrung/Fähigkeiten vorhanden wären
4. "desired"（was HR sehen möchte）：Aus der Perspektive von HR, analysieren Sie, was HR am liebsten sehen möchte

Geben Sie im JSON-Format zurück：
{
  "original": {"summary": "Aktuelle Lebenslauf-Zusammenfassung", "experience": [{"company": "Unternehmen", "position": "Position", "duration": "Zeitraum", "highlights": ["Aktuelle Beschreibung 1", "Beschreibung 2"]}], "skills": ["Aktuelle Fähigkeiten"], "score": 65.0, "analysis": "Objektive Bewertung des aktuellen Lebenslaufs"},
  "optimized": {"summary": "Optimierte Zusammenfassung", "experience": [{"company": "Unternehmen", "position": "Position", "duration": "Zeitraum", "highlights": ["Optimierte Beschreibung 1", "Beschreibung 2"]}], "skills": ["Optimierte Fähigkeiten"], "score": 85.0, "analysis": "Erklärung der Optimierungsstrategie"},
  "imagined": {"summary": "Ideale Zustands-Zusammenfassung", "experience": [{"company": "Unternehmen", "position": "Position", "duration": "Zeitraum", "highlights": ["Ideale Beschreibung 1", "Beschreibung 2"]}], "skills": ["Ideale Fähigkeiten"], "score": 95.0, "analysis": "Wie man den idealen Zustand erreicht"},
  "desired": {"summary": "Was HR sehen möchte", "experience": [{"company": "Unternehmen", "position": "Position", "duration": "Zeitraum", "highlights": ["Was HR sehen möchte 1", "Beschreibung 2"]}], "skills": ["Von HR geschätzte Fähigkeiten"], "score": 88.0, "analysis": "Wichtige HR-Prüfpunkte"}
}`,
	"ar": `أنت محلل سيرة ذاتية محترف. حلل السيرة الذاتية للمستخدم من 4 زوايا.

4 زوايا：
1. "original"（أنا الأصلي）：اعرض محتوى ومستوى السيرة الذاتية الحالي بصدق بدون تجميل
2. "optimized"（أنا المحسّن）：استخدم طريقة STAR + الإنجازات الكمية + كلمات ATS المفتاحية
3. "imagined"（أنا المتخيل）：تخيل الحالة المثالية إذا كان لديك المزيد من الخبرات/المهارات
4. "desired"（ما تريد رؤيته الموارد البشرية）：من منظور الموارد البشرية، حلل ما يريدون رؤيته

أعد بالتنسيق JSON：
{
  "original": {"summary": "ملخص السيرة الذاتية الحالي", "experience": [{"company": "الشركة", "position": "المسمى الوظيفي", "duration": "الفترة", "highlights": ["الوصف الحالي 1", "الوصف 2"]}], "skills": ["المهارات الحالية"], "score": 65.0, "analysis": "تقييم موضوعي للسيرة الذاتية الحالية"},
  "optimized": {"summary": "ملخص محسّن", "experience": [{"company": "الشركة", "position": "المسمى الوظيفي", "duration": "الفترة", "highlights": ["الوصف المحسّن 1", "الوصف 2"]}], "skills": ["المهارات المحسّنة"], "score": 85.0, "analysis": "شرح استراتيجية التحسين"},
  "imagined": {"summary": "ملخص الحالة المثالية", "experience": [{"company": "الشركة", "position": "المسمى الوظيفي", "duration": "الفترة", "highlights": ["الوصف المثالي 1", "الوصف 2"]}], "skills": ["المهارات المثالية"], "score": 95.0, "analysis": "كيف تصل إلى الحالة المثالية"},
  "desired": {"summary": "ما تريد رؤيته الموارد البشرية", "experience": [{"company": "الشركة", "position": "المسمى الوظيفي", "duration": "الفترة", "highlights": ["ما تريد رؤيته الموارد البشرية 1", "الوصف 2"]}], "skills": ["المهارات التي تقدرها الموارد البشرية"], "score": 88.0, "analysis": "نقاط مراجعة الموارد البشرية الرئيسية"}
}`,
	"hi": `आप एक पेशेवर रिज़्यूमे विश्लेषक हैं। उपयोगकर्ता के रिज़्यूमे का 4 दृष्टिकोणों से विश्लेषण करें।

4 दृष्टिकोण：
1. "original"（मूल मैं）：वर्तमान सामग्री और स्तर को सजावट के बिना ईमानदारी से दिखाएं
2. "optimized"（अनुकूलित मैं）：STAR विधि + मात्रात्मक उपलब्धियां + ATS कीवर्ड का उपयोग करें
3. "imagined"（कल्पना किया मैं）：यदि अधिक अनुभव/कौशल होते तो क्या होता, इसकी कल्पना करें
4. "desired"（HR क्या देखना चाहता है）：HR के दृष्टिकोण से विश्लेषण करें कि वे क्या देखना चाहते हैं

JSON प्रारूप में लौटाएं：
{
  "original": {"summary": "वर्तमान रिज़्यूमे सारांश", "experience": [{"company": "कंपनी", "position": "पद", "duration": "अवधि", "highlights": ["वर्तमान विवरण 1", "विवरण 2"]}], "skills": ["वर्तमान कौशल"], "score": 65.0, "analysis": "वर्तमान रिज़्यूमे का वस्तुनिष्ठ मूल्यांकन"},
  "optimized": {"summary": "अनुकूलित सारांश", "experience": [{"company": "कंपनी", "position": "पद", "duration": "अवधि", "highlights": ["अनुकूलित विवरण 1", "विवरण 2"]}], "skills": ["अनुकूलित कौशल"], "score": 85.0, "analysis": "अनुकूलन रणनीति की व्याख्या"},
  "imagined": {"summary": "आदर्श स्थिति सारांश", "experience": [{"company": "कंपनी", "position": "पद", "duration": "अवधि", "highlights": ["आदर्श विवरण 1", "विवरण 2"]}], "skills": ["आदर्श कौशल"], "score": 95.0, "analysis": "आदर्श स्थिति तक कैसे पहुंचें"},
  "desired": {"summary": "HR क्या देखना चाहता है", "experience": [{"company": "कंपनी", "position": "पद", "duration": "अवधि", "highlights": ["HR क्या देखना चाहता है 1", "विवरण 2"]}], "skills": ["HR द्वारा मूल्यांकित कौशल"], "score": 88.0, "analysis": "HR समीक्षा के मुख्य बिंदु"}
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
			Model:   "Qwen/Qwen3-14B",
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ResumeTake/2.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", provider.Name, err.Error())
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
		jsonRegex := regexp.MustCompile(`\{[\s\S]*\}`)
		if match := jsonRegex.FindString(content); match != "" {
			if err2 := json.Unmarshal([]byte(match), &result); err2 != nil {
				return nil, fmt.Errorf("failed to parse %s result", provider.Name)
			}
			return result, nil
		}
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
	userDataPath = getUserDataPath()
	if err := userStore.Load(userDataPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load users: %v\n", err)
	}

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
			"status":           "healthy",
			"timestamp":        time.Now().Format(time.RFC3339),
			"uptime":           time.Since(startTime).String(),
			"requests":         store.Count(),
			"total":            atomic.LoadInt64(&totalRequests),
			"version":          "2.0.0",
			"ai":               providerNames,
			"memory":           fmt.Sprintf("%d MB", getMemUsage()),
			"users":            userStore.Count(),
			"user_persistence": userDataPath != "",
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

	v1.Post("/upload", authRequired, func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "NO_FILE", "message": "No file uploaded"})
		}
		if file.Size <= 0 {
			return c.Status(400).JSON(fiber.Map{"error": "EMPTY_FILE", "message": "File is empty"})
		}
		if file.Size > maxUploadBytes {
			return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "File too large (max 1MB)"})
		}
		filename := filepath.Base(file.Filename)
		if filename != file.Filename || strings.Contains(filename, "\x00") {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_FILENAME", "message": "Invalid filename"})
		}
		ext := strings.ToLower(filepath.Ext(filename))
		if !allowedUploadExt[ext] {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_TYPE", "message": "Only .txt and .md files are supported for safe text parsing"})
		}
		contentType := strings.ToLower(file.Header.Get("Content-Type"))
		allowedMime := strings.HasPrefix(contentType, "text/plain") || contentType == "text/markdown" || contentType == "application/octet-stream"
		if contentType != "" && !allowedMime {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_MIME", "message": "Only plain text and markdown files are supported"})
		}
		f, err := file.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read file"})
		}
		defer f.Close()
		content, err := io.ReadAll(io.LimitReader(f, maxUploadBytes+1))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read file content"})
		}
		if len(content) > maxUploadBytes {
			return c.Status(400).JSON(fiber.Map{"error": "FILE_TOO_LARGE", "message": "File too large (max 1MB)"})
		}
		if bytes.Contains(content, []byte{0}) || !utf8.Valid(content) {
			return c.Status(400).JSON(fiber.Map{"error": "UNSAFE_CONTENT", "message": "Only valid UTF-8 text files are supported"})
		}
		text := string(content)
		text = strings.ReplaceAll(text, "\r\n", "\n")
		text = strings.TrimSpace(text)
		if text == "" {
			return c.Status(400).JSON(fiber.Map{"error": "EMPTY_CONTENT", "message": "Resume text is empty"})
		}
		if len(text) > 15000 {
			text = text[:15000]
		}
		return c.JSON(fiber.Map{
			"success": true,
			"data": map[string]interface{}{
				"filename": filename,
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

	v1.Post("/optimize", authMiddleware, limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), func(c *fiber.Ctx) error {
		user := c.Locals("user").(*User)
		if user.UsageCount >= user.MaxFreeUsage {
			return c.Status(403).JSON(fiber.Map{"error": "LIMIT_EXCEEDED", "message": "Free usage limit exceeded. Please upgrade."})
		}
		userStore.mu.Lock()
		user.UsageCount++
		userStore.mu.Unlock()

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

		return c.JSON(fiber.Map{"success": true, "data": result, "usage_count": user.UsageCount, "max_free_usage": user.MaxFreeUsage})
	})

	v1.Post("/perspective", authMiddleware, limiter.New(limiter.Config{
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

	v1.Post("/auth/send-code", func(c *fiber.Ctx) error {
		var body struct {
			Email string `json:"email"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		body.Email = strings.TrimSpace(strings.ToLower(body.Email))
		if !emailRegex.MatchString(body.Email) {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid email format"})
		}
		if _, exists := userStore.GetByEmail(body.Email); exists {
			return c.Status(409).JSON(fiber.Map{"error": "CONFLICT", "message": "Email already registered"})
		}
		code := generateVerificationCode()
		verificationStore.Save(body.Email, code)
		if err := sendVerificationEmail(body.Email, code); err != nil {
			fmt.Printf("[SMTP] Failed to send verification email: %v\n", err)
		}
		return c.JSON(fiber.Map{"success": true, "message": "Verification code sent"})
	})

	v1.Post("/auth/verify-code", func(c *fiber.Ctx) error {
		var body struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		body.Email = strings.TrimSpace(strings.ToLower(body.Email))
		if verificationStore.Verify(body.Email, body.Code) {
			return c.JSON(fiber.Map{"success": true, "message": "Email verified"})
		}
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_CODE", "message": "Invalid or expired verification code"})
	})

	v1.Post("/auth/register", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		body.Email = strings.TrimSpace(strings.ToLower(body.Email))
		body.Name = strings.TrimSpace(body.Name)
		if !emailRegex.MatchString(body.Email) {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Invalid email format"})
		}
		if len(body.Password) < 6 {
			return c.Status(400).JSON(fiber.Map{"error": "VALIDATION_ERROR", "message": "Password must be at least 6 characters"})
		}
		if body.Name == "" {
			body.Name = body.Email[:strings.Index(body.Email, "@")]
		}
		if _, exists := userStore.GetByEmail(body.Email); exists {
			return c.Status(409).JSON(fiber.Map{"error": "CONFLICT", "message": "Email already registered"})
		}
		hash := sha256.Sum256([]byte(body.Password))
		token := generateToken(body.Email)
		user := &User{
			ID:           uuid.New().String(),
			Email:        body.Email,
			Password:     hex.EncodeToString(hash[:]),
			Name:         body.Name,
			Token:        token,
			MaxFreeUsage: 5,
			CreatedAt:    time.Now(),
		}
		userStore.Save(user)
		return c.Status(201).JSON(fiber.Map{"success": true, "data": user})
	})

	v1.Post("/auth/login", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		body.Email = strings.TrimSpace(strings.ToLower(body.Email))
		user, ok := userStore.GetByEmail(body.Email)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "INVALID_CREDENTIALS", "message": "Invalid email or password"})
		}
		hash := sha256.Sum256([]byte(body.Password))
		if user.Password != hex.EncodeToString(hash[:]) {
			return c.Status(401).JSON(fiber.Map{"error": "INVALID_CREDENTIALS", "message": "Invalid email or password"})
		}
		userStore.mu.Lock()
		user.Token = generateToken(body.Email)
		userStore.mu.Unlock()
		return c.JSON(fiber.Map{"success": true, "data": user})
	})

	v1.Get("/auth/me", authMiddleware, func(c *fiber.Ctx) error {
		user := c.Locals("user").(*User)
		return c.JSON(fiber.Map{"success": true, "data": user})
	})

	v1.Get("/jobs", func(c *fiber.Ctx) error {
		type Job struct {
			ID          string   `json:"id"`
			Title       string   `json:"title"`
			Company     string   `json:"company"`
			Location    string   `json:"location"`
			Salary      string   `json:"salary"`
			Description string   `json:"description"`
			Tags        []string `json:"tags"`
			Type        string   `json:"type"`
			PostedAt    string   `json:"posted_at"`
		}
		jobs := []Job{
			{ID: "j001", Title: "Senior AI Engineer", Company: "ByteDance", Location: "Beijing, China", Salary: "40k-70k", Description: "Build large language model applications and AI-powered products for millions of users.", Tags: []string{"Python", "LLM", "PyTorch", "RAG"}, Type: "full-time", PostedAt: "2026-07-01"},
			{ID: "j002", Title: "Frontend Engineer", Company: "Alibaba Cloud", Location: "Hangzhou, China", Salary: "30k-55k", Description: "Develop next-generation cloud console and developer tools.", Tags: []string{"React", "TypeScript", "Ant Design", "Node.js"}, Type: "full-time", PostedAt: "2026-06-28"},
			{ID: "j003", Title: "Backend Developer", Company: "Tencent", Location: "Shenzhen, China", Salary: "35k-60k", Description: "Design and implement high-performance microservices for WeChat ecosystem.", Tags: []string{"Go", "gRPC", "Redis", "MySQL"}, Type: "full-time", PostedAt: "2026-06-25"},
			{ID: "j004", Title: "Product Manager", Company: "Meituan", Location: "Beijing, China", Salary: "30k-50k", Description: "Lead product strategy for local life services platform.", Tags: []string{"Product Strategy", "Data Analysis", "Agile", "User Research"}, Type: "full-time", PostedAt: "2026-07-02"},
			{ID: "j005", Title: "UI/UX Designer", Company: "Xiaomi", Location: "Beijing, China", Salary: "25k-45k", Description: "Design intuitive interfaces for MIUI and smart home products.", Tags: []string{"Figma", "Prototyping", "Design System", "Interaction Design"}, Type: "full-time", PostedAt: "2026-06-30"},
			{ID: "j006", Title: "ML Engineer Intern", Company: "Baidu", Location: "Beijing, China", Salary: "4k-8k", Description: "Work on cutting-edge NLP and computer vision projects.", Tags: []string{"Python", "TensorFlow", "NLP", "Computer Vision"}, Type: "intern", PostedAt: "2026-07-03"},
			{ID: "j007", Title: "Software Engineer", Company: "Google", Location: "Mountain View, CA", Salary: "$150k-$220k", Description: "Build scalable infrastructure and services for Google Cloud.", Tags: []string{"Java", "C++", "Distributed Systems", "GCP"}, Type: "full-time", PostedAt: "2026-06-20"},
			{ID: "j008", Title: "Frontend Developer", Company: "Microsoft", Location: "Redmond, WA", Salary: "$130k-$190k", Description: "Develop Azure portal features and React component libraries.", Tags: []string{"React", "TypeScript", "Azure", "WCAG"}, Type: "full-time", PostedAt: "2026-06-22"},
			{ID: "j009", Title: "Full Stack Engineer", Company: "Shopify", Location: "Remote (Global)", Salary: "$120k-$175k", Description: "Build merchant-facing tools and Ruby on Rails applications.", Tags: []string{"Ruby on Rails", "React", "GraphQL", "PostgreSQL"}, Type: "full-time", PostedAt: "2026-07-01"},
			{ID: "j010", Title: "Data Scientist Intern", Company: "Netflix", Location: "Los Gatos, CA", Salary: "$6k-$10k/mo", Description: "Analyze user engagement data and build recommendation models.", Tags: []string{"Python", "SQL", "Spark", "A/B Testing"}, Type: "intern", PostedAt: "2026-06-27"},
			{ID: "j011", Title: "DevOps Engineer", Company: "Huawei", Location: "Shenzhen, China", Salary: "30k-50k", Description: "Maintain CI/CD pipelines and cloud-native infrastructure.", Tags: []string{"Kubernetes", "Docker", "Jenkins", "Linux"}, Type: "full-time", PostedAt: "2026-06-29"},
			{ID: "j012", Title: "iOS Developer", Company: "ByteDance", Location: "Shanghai, China", Salary: "30k-55k", Description: "Build TikTok iOS features used by billions worldwide.", Tags: []string{"Swift", "Objective-C", "UIKit", "Core Animation"}, Type: "full-time", PostedAt: "2026-07-02"},
			{ID: "j013", Title: "AI Research Intern", Company: "OpenAI", Location: "San Francisco, CA", Salary: "$8k-$12k/mo", Description: "Contribute to frontier AI safety and alignment research.", Tags: []string{"Python", "PyTorch", "RLHF", "Research"}, Type: "intern", PostedAt: "2026-06-26"},
			{ID: "j014", Title: "Backend Engineer", Company: "Stripe", Location: "Remote (US)", Salary: "$140k-$200k", Description: "Build reliable payment infrastructure and APIs.", Tags: []string{"Ruby", "Go", "Distributed Systems", "API Design"}, Type: "full-time", PostedAt: "2026-07-03"},
			{ID: "j015", Title: "Product Designer", Company: "Alibaba", Location: "Hangzhou, China", Salary: "25k-45k", Description: "Design end-to-end experiences for Taobao merchants.", Tags: []string{"Figma", "User Research", "Service Design", "Data Visualization"}, Type: "full-time", PostedAt: "2026-06-30"},
			{ID: "j016", Title: "QA Automation Engineer", Company: "JD.com", Location: "Beijing, China", Salary: "20k-35k", Description: "Build automated testing frameworks for e-commerce platform.", Tags: []string{"Selenium", "Jenkins", "Python", "API Testing"}, Type: "full-time", PostedAt: "2026-06-28"},
			{ID: "j017", Title: "Cloud Solutions Architect", Company: "AWS", Location: "Seattle, WA", Salary: "$160k-$230k", Description: "Design enterprise cloud architectures and migration strategies.", Tags: []string{"AWS", "Terraform", "Microservices", "Security"}, Type: "full-time", PostedAt: "2026-06-24"},
			{ID: "j018", Title: "React Native Developer", Company: "Shopee", Location: "Singapore", Salary: "SGD 5k-9k", Description: "Build cross-platform mobile features for Southeast Asian market.", Tags: []string{"React Native", "TypeScript", "Redux", "REST API"}, Type: "full-time", PostedAt: "2026-07-01"},
			{ID: "j019", Title: "Cybersecurity Analyst Intern", Company: "Kaspersky", Location: "Moscow / Remote", Salary: "$3k-$5k/mo", Description: "Analyze threat intelligence and malware samples.", Tags: []string{"Python", "SIEM", "Threat Analysis", "Reverse Engineering"}, Type: "intern", PostedAt: "2026-06-25"},
			{ID: "j020", Title: "Blockchain Developer", Company: "Ant Group", Location: "Shanghai, China", Salary: "35k-60k", Description: "Build Web3 and blockchain-based financial products.", Tags: []string{"Solidity", "Go", "Hyperledger", "Smart Contracts"}, Type: "full-time", PostedAt: "2026-07-02"},
			{ID: "j021", Title: "Embedded Systems Engineer", Company: "NIO", Location: "Shanghai, China", Salary: "25k-45k", Description: "Develop real-time embedded software for autonomous driving systems.", Tags: []string{"C/C++", "RTOS", "Linux Kernel", "CAN Protocol"}, Type: "full-time", PostedAt: "2026-06-27"},
			{ID: "j022", Title: "Technical Writer", Company: "Atlassian", Location: "Remote (Global)", Salary: "$90k-$130k", Description: "Write developer documentation and API references for Jira and Confluence.", Tags: []string{"Markdown", "OpenAPI", "Git", "Technical Writing"}, Type: "full-time", PostedAt: "2026-06-29"},
		}
		return c.JSON(fiber.Map{"success": true, "data": jobs})
	})

	v1.Get("/jobs/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{
			"id":          id,
			"title":       "Senior AI Engineer",
			"company":     "ByteDance",
			"location":    "Beijing, China",
			"salary":      "40k-70k",
			"description": "Build large language model applications and AI-powered products. You will work closely with research teams to productionize cutting-edge models.",
			"requirements": []string{
				"3+ years of ML engineering experience",
				"Strong Python and PyTorch skills",
				"Experience with LLM fine-tuning and deployment",
				"Familiarity with RAG and vector databases",
			},
			"tags":      []string{"Python", "LLM", "PyTorch", "RAG"},
			"type":      "full-time",
			"posted_at": "2026-07-01",
		}})
	})

	v1.Post("/generate-resume", authMiddleware, limiter.New(limiter.Config{
		Max:        10,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}), func(c *fiber.Ctx) error {
		if os.Getenv("ENABLE_GENERATE_RESUME") != "true" {
			return c.Status(402).JSON(fiber.Map{
				"success": false,
				"error":   "PAYMENT_REQUIRED",
				"message": "Zero-basis resume generation is a paid feature and is temporarily disabled while payment is being configured.",
			})
		}
		var body struct {
			Messages []GroqMessage `json:"messages"`
			Lang     string        `json:"lang"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		if len(body.Messages) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "NO_MESSAGES", "message": "Messages array is required"})
		}
		if body.Lang == "" {
			body.Lang = "en"
		}

		systemPrompt := map[string]string{
			"zh": `你是一位专业的零基础简历生成助手。你将通过对话的方式，一步一步引导用户构建一份完整的简历。

工作流程：
1. 首先询问：姓名、目标职位、联系方式（邮箱/电话/城市）
2. 然后询问：教育背景（学校、专业、学位、时间）
3. 接着询问：工作经历（公司、职位、时间、主要成就，引导用户用STAR法则描述）
4. 再询问：项目经验（项目名称、角色、技术栈、成果）
5. 然后询问：技能清单（技术技能、软技能）
6. 最后询问：其他信息（证书、语言、兴趣爱好等）

规则：
- 每次只问一个问题类别，不要一次问太多
- 用友好、鼓励的语气引导用户
- 如果用户回答过于简单，主动追问细节
- 当收集到足够信息后，生成完整的JSON格式简历
- 最终简历JSON格式：{"resume": {"name": "", "title": "", "contact": {"email": "", "phone": "", "location": ""}, "summary": "", "education": [], "experience": [], "projects": [], "skills": [], "other": ""}}

请用中文与用户交流。`,
			"en": `You are a professional zero-basis resume generation assistant. You will guide users step by step to build a complete resume through conversation.

Workflow:
1. First ask: name, target position, contact info (email/phone/city)
2. Then ask: education background (school, major, degree, dates)
3. Then ask: work experience (company, role, dates, key achievements using STAR method)
4. Then ask: project experience (project name, role, tech stack, outcomes)
5. Then ask: skills list (technical skills, soft skills)
6. Finally ask: any other info (certifications, languages, hobbies)

Rules:
- Ask one category at a time, don't overwhelm the user
- Use a friendly, encouraging tone
- If user answers are too brief, proactively ask for more details
- When enough information is collected, generate a complete JSON resume
- Final resume JSON format: {"resume": {"name": "", "title": "", "contact": {"email": "", "phone": "", "location": ""}, "summary": "", "education": [], "experience": [], "projects": [], "skills": [], "other": ""}}

Communicate with the user in English.`,
		}

		prompt, ok := systemPrompt[body.Lang]
		if !ok {
			prompt = systemPrompt["en"]
		}

		providers := getAIProviders()
		if len(providers) == 0 {
			return c.Status(503).JSON(fiber.Map{"error": "NO_AI", "message": "No AI provider configured"})
		}

		messages := []GroqMessage{{Role: "system", Content: prompt}}
		messages = append(messages, body.Messages...)

		var lastErr error
		for _, p := range providers {
			reqBody := GroqRequest{
				Model:       p.Model,
				Messages:    messages,
				MaxTokens:   2048,
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
			response := fiber.Map{"message": content}
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(content), &parsed); err == nil {
				if _, hasResume := parsed["resume"]; hasResume {
					response["resume_complete"] = true
					response["resume"] = parsed["resume"]
				}
			} else {
				jsonRegex := regexp.MustCompile(`\{[\s\S]*"resume"[\s\S]*\}`)
				if match := jsonRegex.FindString(content); match != "" {
					if err2 := json.Unmarshal([]byte(match), &parsed); err2 == nil {
						if _, hasResume := parsed["resume"]; hasResume {
							response["resume_complete"] = true
							response["resume"] = parsed["resume"]
						}
					}
				}
			}
			return c.JSON(fiber.Map{"success": true, "data": response})
		}
		return c.Status(503).JSON(fiber.Map{"success": false, "error": lastErr.Error()})
	})

	v1.Get("/pricing", func(c *fiber.Ctx) error {
		tiers := []fiber.Map{
			{
				"id":    "free",
				"name":  "Free",
				"price": 0,
				"features": []string{
					"5 AI optimizations per month",
					"Basic resume templates",
					"1 language support",
					"Download as text",
				},
				"usage_limit": 5,
			},
			{
				"id":    "pro",
				"name":  "Pro",
				"price": 9.9,
				"features": []string{
					"Unlimited AI optimizations",
					"All premium templates",
					"10+ language support",
					"PDF export",
					"ATS score analysis",
					"Perspective analysis",
					"Priority support",
				},
				"usage_limit": -1,
			},
			{
				"id":    "enterprise",
				"name":  "Enterprise",
				"price": 49.9,
				"features": []string{
					"Everything in Pro",
					"Team collaboration",
					"Custom templates",
					"API access",
					"Batch resume processing",
					"Dedicated support",
					"SLA guarantee",
				},
				"usage_limit": -1,
			},
		}
		return c.JSON(fiber.Map{"success": true, "data": tiers})
	})

	v1.Post("/create-checkout-session", authMiddleware, func(c *fiber.Ctx) error {
		stripeKey := os.Getenv("STRIPE_SECRET_KEY")
		if stripeKey == "" {
			return c.Status(503).JSON(fiber.Map{"error": "PAYMENT_NOT_CONFIGURED", "message": "Payment system not configured"})
		}
		var body struct {
			Plan string `json:"plan"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
		}
		user := c.Locals("user").(*User)
		priceMap := map[string]string{
			"pro":        "price_pro_monthly_29",
			"enterprise": "price_enterprise_monthly_99",
		}
		priceID, ok := priceMap[body.Plan]
		if !ok {
			return c.Status(400).JSON(fiber.Map{"error": "INVALID_PLAN", "message": "Invalid plan"})
		}
		payload := fmt.Sprintf(`{
			"mode": "subscription",
			"payment_method_types": ["card", "alipay", "wechat_pay"],
			"customer_email": "%s",
			"line_items": [{"price": "%s", "quantity": 1}],
			"success_url": "https://resume.takee.top/%s/pricing?session_id={CHECKOUT_SESSION_ID}",
			"cancel_url": "https://resume.takee.top/%s/pricing",
			"metadata": {"user_id": "%s", "plan": "%s"}
		}`, user.Email, priceID, "en", "en", user.ID, body.Plan)
		req, err := http.NewRequest("POST", "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(payload))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "INTERNAL", "message": "Failed to create checkout session"})
		}
		req.Header.Set("Authorization", "Bearer "+stripeKey)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{"error": "STRIPE_ERROR", "message": "Failed to connect to Stripe"})
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "INTERNAL", "message": "Invalid Stripe response"})
		}
		if resp.StatusCode != 200 {
			return c.Status(502).JSON(fiber.Map{"error": "STRIPE_ERROR", "message": result["error"]})
		}
		return c.JSON(fiber.Map{"success": true, "url": result["url"]})
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
