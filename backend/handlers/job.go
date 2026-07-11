package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/transform"

	"resumetake/services"
)

type JobHandler struct{}

var jobHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		// Custom DialContext checks the resolved IP at connection time to
		// prevent DNS rebinding attacks (where the first lookup returns a
		// public IP to pass validation, then the actual connection resolves
		// to a private IP).
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			// Validate ALL resolved IPs, then dial the first valid one
			// directly (not the hostname) to prevent DNS rebinding: if we
			// passed the hostname to the dialer, it would re-resolve and
			// could get a different (blocked) IP on the second lookup.
			var validIP net.IP
			for _, ip := range ips {
				if isBlockedIP(ip.IP) {
					return nil, fmt.Errorf("blocked address: %s", ip.IP)
				}
				if validIP == nil {
					validIP = ip.IP
				}
			}
			if validIP == nil {
				return nil, fmt.Errorf("no valid IP resolved for %s", host)
			}
			dialer := &net.Dialer{Timeout: 10 * time.Second}
			// Dial the IP directly — TLS SNI/cert validation is handled by
			// http.Transport using the request URL's hostname, not addr.
			return dialer.DialContext(ctx, network, net.JoinHostPort(validIP.String(), port))
		},
		IdleConnTimeout: 90 * time.Second, // R58-B2: close idle conns (default 0 = no limit)
		MaxIdleConns:    20,
	},
	// CheckRedirect re-validates every redirect URL to prevent SSRF via
	// redirects (e.g. a public URL that 302s to http://127.0.0.1/...).
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		// R52b-B2: only allow HTTPS redirects to match the initial URL policy
		// (line ~128-130 upgrades initial URL to HTTPS). Allowing http://
		// redirects would let the server fetch content over plaintext and be
		// used as an open HTTP fetcher.
		// R52b-B3: strip userinfo from redirect URLs — a redirecting server
		// could redirect to https://user:pass@host/ causing Go's http.Client
		// to send Basic Auth to the target.
		parsed, err := url.Parse(req.URL.String())
		if err != nil || parsed.Scheme != "https" {
			return fmt.Errorf("invalid redirect URL")
		}
		parsed.User = nil
		req.URL.User = nil
		ips, err := net.DefaultResolver.LookupIPAddr(req.Context(), parsed.Hostname())
		if err != nil {
			return fmt.Errorf("failed to resolve redirect host")
		}
		for _, ip := range ips {
			if isBlockedIP(ip.IP) {
				return fmt.Errorf("redirect to blocked address")
			}
		}
		return nil
	},
}

func NewJobHandler() *JobHandler {
	return &JobHandler{}
}

// isBlockedHost 检查解析后的 IP 是否为内网/保留地址，防止 SSRF
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	// IPv4 映射的 IPv6
	if v4 := ip.To4(); v4 != nil {
		// 0.0.0.0/8 "This Network" — on Linux, traffic to 0.0.0.1 etc.
		// can be routed to loopback, bypassing IsLoopback() checks.
		if v4[0] == 0 {
			return true
		}
		if v4[0] == 169 && v4[1] == 254 {
			return true // 云元数据 169.254.x.x
		}
		// CGNAT 100.64.0.0/10 (RFC 6598) — some cloud environments run
		// internal services in this range. Go's IsPrivate does not cover it.
		if v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
			return true
		}
	}
	return false
}

// decodeHTMLBody decodes the response body to UTF-8 based on the charset
// declared in the Content-Type header or HTML meta tags.
// R57-B1: previously ScrapeJob did string(bodyBytes) which treats the bytes
// as UTF-8 — GBK/GB2312 encoded Chinese job pages returned garbled mojibake,
// making the scraped title/description/text unusable for AI processing.
// Most modern sites use UTF-8, so we fast-path valid UTF-8 with zero overhead.
func decodeHTMLBody(bodyBytes []byte, contentType string) string {
	// Fast path: if already valid UTF-8, no decoding needed (common case).
	if utf8.Valid(bodyBytes) {
		return string(bodyBytes)
	}

	charset := detectCharset(contentType, bodyBytes)
	if charset == "" || strings.EqualFold(charset, "utf-8") || strings.EqualFold(charset, "utf8") {
		// Unknown charset but not valid UTF-8 — return as-is (best effort).
		return string(bodyBytes)
	}

	enc, err := htmlindex.Get(charset)
	if err != nil {
		// Unsupported charset label — return raw bytes as-is.
		return string(bodyBytes)
	}
	decoded, _, err := transform.Bytes(enc.NewDecoder(), bodyBytes)
	if err != nil {
		return string(bodyBytes)
	}
	return string(decoded)
}

// detectCharset extracts the charset from the Content-Type header first,
// falling back to the HTML <meta charset=...> or <meta http-equiv> tag.
func detectCharset(contentType string, bodyBytes []byte) string {
	// 1. Content-Type header: "text/html; charset=gbk"
	ct := strings.ToLower(contentType)
	if idx := strings.Index(ct, "charset="); idx != -1 {
		cs := strings.TrimSpace(ct[idx+8:])
		// Strip trailing params (e.g. "; boundary=...")
		if semi := strings.IndexByte(cs, ';'); semi != -1 {
			cs = cs[:semi]
		}
		cs = strings.Trim(cs, `"' `)
		if cs != "" {
			return cs
		}
	}

	// 2. HTML meta tag — only scan the first 2KB (charset declaration is
	// always near the top of the document, before <body>).
	head := bodyBytes
	if len(head) > 2048 {
		head = head[:2048]
	}
	headLower := strings.ToLower(string(head))
	// <meta charset="gbk">
	if idx := strings.Index(headLower, "<meta charset="); idx != -1 {
		rest := headLower[idx+14:]
		cs := strings.TrimLeft(rest, " ='\"")
		end := strings.IndexAny(cs, " >\"'")
		if end > 0 {
			return cs[:end]
		}
	}
	// <meta http-equiv="Content-Type" content="text/html; charset=gbk">
	if idx := strings.Index(headLower, "http-equiv=\"content-type\""); idx != -1 {
		// Find the content= attribute after this position.
		contentArea := headLower[idx:]
		if ci := strings.Index(contentArea, "charset="); ci != -1 {
			cs := strings.TrimSpace(contentArea[ci+8:])
			if semi := strings.IndexByte(cs, ';'); semi != -1 {
				cs = cs[:semi]
			}
			cs = strings.Trim(cs, `"' `)
			if cs != "" {
				return cs
			}
		}
	}
	return ""
}

func (h *JobHandler) ScrapeJob(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_BODY", "message": "Invalid request body"})
	}

	jobURL, _ := body["url"].(string)
	if jobURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "NO_URL", "message": "URL is required"})
	}
	if len(jobURL) > 2048 {
		return c.Status(400).JSON(fiber.Map{"error": "URL_TOO_LONG", "message": "URL exceeds maximum length (2048 characters)"})
	}

	// R31-4: force HTTPS only — accepting http:// sends the fetched job
	// content over plaintext, exposing it to MITM. Most job sites support
	// HTTPS; upgrade http:// automatically and reject if HTTPS isn't used.
	// R34-L5: reject non-http(s) schemes explicitly — blindly prepending
	// "https://" to "ftp://example.com" produces "https://ftp://example.com",
	// a confusing malformed URL that fails later with an opaque error.
	// R53-B2: scheme check is case-insensitive per RFC 3986 — "HTTP://"
	// is a valid scheme but was previously rejected.
	lowerURL := strings.ToLower(jobURL[:min(len(jobURL), 8)])
	if strings.Contains(jobURL, "://") && !strings.HasPrefix(lowerURL, "http://") && !strings.HasPrefix(lowerURL, "https://") {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_URL", "message": "Only HTTP and HTTPS URLs are supported"})
	}
	if strings.HasPrefix(lowerURL, "http://") {
		jobURL = "https://" + jobURL[len("http://"):]
	} else if !strings.HasPrefix(lowerURL, "https://") {
		jobURL = "https://" + jobURL
	}

	parsedURL, err := url.Parse(jobURL)
	if err != nil || parsedURL.Scheme != "https" {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_URL", "message": "Invalid URL"})
	}

	// R53-B1: strip userinfo from URL — "https://user:pass@host/" causes
	// net/http to send Basic Auth headers to the target, allowing this
	// server to be abused as a credential-stuffing relay.
	parsedURL.User = nil
	jobURL = parsedURL.String()

	// SSRF 防护：解析主机名并检查 IP（带 context，防止慢速 DNS 阻塞）
	host := parsedURL.Hostname()
	ips, err := net.DefaultResolver.LookupIPAddr(c.Context(), host)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_URL", "message": "Failed to resolve host"})
	}
	for _, ip := range ips {
		if isBlockedIP(ip.IP) {
			return c.Status(400).JSON(fiber.Map{"error": "BLOCKED_URL", "message": "URL points to a blocked address"})
		}
	}

	req, err := http.NewRequestWithContext(c.Context(), "GET", jobURL, nil)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "INVALID_URL", "message": "Invalid URL"})
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := jobHTTPClient.Do(req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "FETCH_FAILED", "message": "Failed to fetch job page"})
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// Don't echo the upstream status code to the client — it leaks
		// internal fetch behavior and lets attackers probe target sites.
		// R40-L3: strip query string from logged URL — job posting URLs
		// may contain sensitive params (tokens, session ids, ref codes).
		safeURL := jobURL
		if idx := strings.Index(safeURL, "?"); idx >= 0 {
			safeURL = safeURL[:idx]
		}
		log.Printf("[WARN] ScrapeJob upstream returned HTTP %d for %s", resp.StatusCode, safeURL)
		return c.Status(502).JSON(fiber.Map{"error": "FETCH_FAILED", "message": "Failed to fetch job page"})
	}

	// R43-B2: Validate Content-Type BEFORE reading body — non-HTML responses
	// (PDF, JSON, images, video) would waste up to 500KB of bandwidth and
	// memory before being rejected. Check headers first, read body only if
	// the content type is acceptable.
	// R51-B4: lowercase the Content-Type before substring check — HTTP
	// header values are case-insensitive and some servers return "Text/Html".
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType != "" &&
		!strings.Contains(contentType, "text/html") &&
		!strings.Contains(contentType, "application/xhtml") &&
		!strings.Contains(contentType, "text/plain") {
		return c.Status(422).JSON(fiber.Map{"error": "UNSUPPORTED_CONTENT_TYPE", "message": "Job page must be an HTML page"})
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 500*1024))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "READ_ERROR", "message": "Failed to read page"})
	}

	html := decodeHTMLBody(bodyBytes, contentType)
	title := services.ExtractMeta(html, "og:title")
	if title == "" {
		// R57-B6: fall back to <title> tag — many sites (especially Chinese
		// job boards) don't implement Open Graph but always set <title>.
		title = services.ExtractTitle(html)
	}
	desc := services.ExtractMeta(html, "description")
	if desc == "" {
		desc = services.ExtractMeta(html, "og:description")
	}

	cleanText := services.StripHTML(html)
	if len(cleanText) > 5000 {
		cleanText = services.TruncateUTF8(cleanText, 5000)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": map[string]interface{}{
			"url":         jobURL,
			"title":       title,
			"description": desc,
			"text":        cleanText,
		},
	})
}

// Job represents a job listing returned by the API.
type Job struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Location    string   `json:"location"`
	Salary      string   `json:"salary"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Type        string   `json:"type"`
	URL         string   `json:"url"`
	PostedAt    string   `json:"posted_at"`
}

// jobListings is the static demo dataset shared by GetJobs and GetJob.
var jobListings = []Job{
	{ID: "j001", Title: "Senior AI Engineer", Company: "ByteDance", Location: "Beijing, China", Salary: "40k-70k", Description: "Build large language model applications and AI-powered products for millions of users.", Tags: []string{"Python", "LLM", "PyTorch", "RAG"}, Type: "full-time", URL: "https://jobs.bytedance.com/en/position?keywords=&category=&location=&project=&type=&job_hot_flag=&current=1&limit=10", PostedAt: "2026-07-01"},
	{ID: "j002", Title: "Frontend Engineer", Company: "Alibaba Cloud", Location: "Hangzhou, China", Salary: "30k-55k", Description: "Develop next-generation cloud console and developer tools.", Tags: []string{"React", "TypeScript", "Ant Design", "Node.js"}, Type: "full-time", URL: "https://talent.alibaba.com/position-list?lang=zh", PostedAt: "2026-06-28"},
	{ID: "j003", Title: "Backend Developer", Company: "Tencent", Location: "Shenzhen, China", Salary: "35k-60k", Description: "Design and implement high-performance microservices for WeChat ecosystem.", Tags: []string{"Go", "gRPC", "Redis", "MySQL"}, Type: "full-time", URL: "https://careers.tencent.com/en-us/positions.html", PostedAt: "2026-06-25"},
	{ID: "j004", Title: "Product Manager", Company: "Meituan", Location: "Beijing, China", Salary: "30k-50k", Description: "Lead product strategy for local life services platform.", Tags: []string{"Product Strategy", "Data Analysis", "Agile", "User Research"}, Type: "full-time", URL: "https://zhaopin.meituan.com/", PostedAt: "2026-07-02"},
	{ID: "j005", Title: "UI/UX Designer", Company: "Xiaomi", Location: "Beijing, China", Salary: "25k-45k", Description: "Design intuitive interfaces for MIUI and smart home products.", Tags: []string{"Figma", "Prototyping", "Design System", "Interaction Design"}, Type: "full-time", URL: "https://hr.xiaomi.com/campus", PostedAt: "2026-06-30"},
	{ID: "j006", Title: "ML Engineer Intern", Company: "Baidu", Location: "Beijing, China", Salary: "4k-8k", Description: "Work on cutting-edge NLP and computer vision projects.", Tags: []string{"Python", "TensorFlow", "NLP", "Computer Vision"}, Type: "intern", URL: "https://talent.baidu.com/external/baidu/campus.html", PostedAt: "2026-07-03"},
	{ID: "j007", Title: "Software Engineer", Company: "Google", Location: "Mountain View, CA", Salary: "$150k-$220k", Description: "Build scalable infrastructure and services for Google Cloud.", Tags: []string{"Java", "C++", "Distributed Systems", "GCP"}, Type: "full-time", URL: "https://careers.google.com/jobs/results/?q=software%20engineer", PostedAt: "2026-06-20"},
	{ID: "j008", Title: "Frontend Developer", Company: "Microsoft", Location: "Redmond, WA", Salary: "$130k-$190k", Description: "Develop Azure portal features and React component libraries.", Tags: []string{"React", "TypeScript", "Azure", "WCAG"}, Type: "full-time", URL: "https://jobs.careers.microsoft.com/global/en/search?q=frontend%20developer", PostedAt: "2026-06-22"},
	{ID: "j009", Title: "Full Stack Engineer", Company: "Shopify", Location: "Remote (Global)", Salary: "$120k-$175k", Description: "Build merchant-facing tools and Ruby on Rails applications.", Tags: []string{"Ruby on Rails", "React", "GraphQL", "PostgreSQL"}, Type: "full-time", URL: "https://www.shopify.com/careers/search", PostedAt: "2026-07-01"},
	{ID: "j010", Title: "Data Scientist Intern", Company: "Netflix", Location: "Los Gatos, CA", Salary: "$6k-$10k/mo", Description: "Analyze user engagement data and build recommendation models.", Tags: []string{"Python", "SQL", "Spark", "A/B Testing"}, Type: "intern", URL: "https://jobs.netflix.com/search?q=data%20scientist", PostedAt: "2026-06-27"},
	{ID: "j011", Title: "DevOps Engineer", Company: "Huawei", Location: "Shenzhen, China", Salary: "30k-50k", Description: "Maintain CI/CD pipelines and cloud-native infrastructure.", Tags: []string{"Kubernetes", "Docker", "Jenkins", "Linux"}, Type: "full-time", URL: "https://career.huawei.com/reccampportal/portal5/index.html", PostedAt: "2026-06-29"},
	{ID: "j012", Title: "iOS Developer", Company: "ByteDance", Location: "Shanghai, China", Salary: "30k-55k", Description: "Build TikTok iOS features used by billions worldwide.", Tags: []string{"Swift", "Objective-C", "UIKit", "Core Animation"}, Type: "full-time", URL: "https://jobs.bytedance.com/en/position?keywords=&category=&location=&project=&type=&job_hot_flag=&current=1&limit=10", PostedAt: "2026-07-02"},
	{ID: "j013", Title: "AI Research Intern", Company: "OpenAI", Location: "San Francisco, CA", Salary: "$8k-$12k/mo", Description: "Contribute to frontier AI safety and alignment research.", Tags: []string{"Python", "PyTorch", "RLHF", "Research"}, Type: "intern", URL: "https://openai.com/careers/search?q=research+intern", PostedAt: "2026-06-26"},
	{ID: "j014", Title: "Backend Engineer", Company: "Stripe", Location: "Remote (US)", Salary: "$140k-$200k", Description: "Build reliable payment infrastructure and APIs.", Tags: []string{"Ruby", "Go", "Distributed Systems", "API Design"}, Type: "full-time", URL: "https://stripe.com/jobs/search?query=backend+engineer", PostedAt: "2026-07-03"},
	{ID: "j015", Title: "Product Designer", Company: "Alibaba", Location: "Hangzhou, China", Salary: "25k-45k", Description: "Design end-to-end experiences for Taobao merchants.", Tags: []string{"Figma", "User Research", "Service Design", "Data Visualization"}, Type: "full-time", URL: "https://talent.alibaba.com/position-list?lang=zh", PostedAt: "2026-06-30"},
	{ID: "j016", Title: "QA Automation Engineer", Company: "JD.com", Location: "Beijing, China", Salary: "20k-35k", Description: "Build automated testing frameworks for e-commerce platform.", Tags: []string{"Selenium", "Jenkins", "Python", "API Testing"}, Type: "full-time", URL: "https://campus.jd.com/#/jobs", PostedAt: "2026-06-28"},
	{ID: "j017", Title: "Cloud Solutions Architect", Company: "AWS", Location: "Seattle, WA", Salary: "$160k-$230k", Description: "Design enterprise cloud architectures and migration strategies.", Tags: []string{"AWS", "Terraform", "Microservices", "Security"}, Type: "full-time", URL: "https://www.amazon.jobs/en/search?offset=0&result_limit=10&sort=relevant&category=software-ec2&city=&region=&country=&loc_group_id=&invalid_location=false&country=&loc_group_id=&search_type=keyword&query=cloud+solutions+architect", PostedAt: "2026-06-24"},
	{ID: "j018", Title: "React Native Developer", Company: "Shopee", Location: "Singapore", Salary: "SGD 5k-9k", Description: "Build cross-platform mobile features for Southeast Asian market.", Tags: []string{"React Native", "TypeScript", "Redux", "REST API"}, Type: "full-time", URL: "https://careers.shopee.sg/apply/", PostedAt: "2026-07-01"},
	{ID: "j019", Title: "Cybersecurity Analyst Intern", Company: "Kaspersky", Location: "Moscow / Remote", Salary: "$3k-$5k/mo", Description: "Analyze threat intelligence and malware samples.", Tags: []string{"Python", "SIEM", "Threat Analysis", "Reverse Engineering"}, Type: "intern", URL: "https://www.kaspersky.com/careers/internships", PostedAt: "2026-06-25"},
	{ID: "j020", Title: "Blockchain Developer", Company: "Ant Group", Location: "Shanghai, China", Salary: "35k-60k", Description: "Build Web3 and blockchain-based financial products.", Tags: []string{"Solidity", "Go", "Hyperledger", "Smart Contracts"}, Type: "full-time", URL: "https://talent.antgroup.com/off-campus-position?lang=zh", PostedAt: "2026-07-02"},
	{ID: "j021", Title: "Embedded Systems Engineer", Company: "NIO", Location: "Shanghai, China", Salary: "25k-45k", Description: "Develop real-time embedded software for autonomous driving systems.", Tags: []string{"C/C++", "RTOS", "Linux Kernel", "CAN Protocol"}, Type: "full-time", URL: "https://jobs.nio.com/", PostedAt: "2026-06-27"},
	{ID: "j022", Title: "Technical Writer", Company: "Atlassian", Location: "Remote (Global)", Salary: "$90k-$130k", Description: "Write developer documentation and API references for Jira and Confluence.", Tags: []string{"Markdown", "OpenAPI", "Git", "Technical Writing"}, Type: "full-time", URL: "https://www.atlassian.com/company/careers/all-roles", PostedAt: "2026-06-29"},
}

func (h *JobHandler) GetJobs(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "data": jobListings})
}

func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	id := c.Params("id")
	for _, job := range jobListings {
		if job.ID == id {
			return c.JSON(fiber.Map{"success": true, "data": job})
		}
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error":   "NOT_FOUND",
		"message": "Job not found",
	})
}
