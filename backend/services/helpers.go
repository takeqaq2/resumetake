package services

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TruncateUTF8 truncates s to at most maxBytes bytes, ensuring the result is
// valid UTF-8 by only keeping complete runes. A naive s[:maxBytes] can split
// a multi-byte rune (Chinese/Japanese/Arabic etc.), producing invalid UTF-8
// that causes parsing errors downstream (AI providers, JSON, etc.).
func TruncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Walk runes, accumulating bytes, until the next rune would exceed maxBytes.
	bytes := 0
	for bytes < len(s) {
		_, size := utf8.DecodeRuneInString(s[bytes:])
		if bytes+size > maxBytes {
			break
		}
		bytes += size
	}
	return s[:bytes]
}

// R57b-B4: renamed parameter from `html` to `htmlContent` to avoid
// shadowing the `html` standard library package. Under Go 1.25 (Docker),
// the shadowing caused `html.UnescapeString` in ExtractTitle (line 200)
// to fail compilation with "html.UnescapeString undefined (type string
// has no field or method UnescapeString)". Go 1.26 (local) was lenient.
func ExtractMeta(htmlContent, name string) string {
	// Anchor on name=" (or property=") to avoid matching substrings inside
	// other attribute names — e.g. searching "description" would otherwise
	// match inside "og:description" and return the wrong content.
	idx := strings.Index(htmlContent, `name="`+name+`"`)
	if idx == -1 {
		idx = strings.Index(htmlContent, `property="`+name+`"`)
	}
	if idx == -1 {
		return ""
	}
	// R50-B2: the content= attribute may appear BEFORE name= in the tag
	// (e.g. <meta content="..." name="description">). Previously we only
	// searched forward from name=, missing the content entirely. Now find
	// the enclosing <meta ...> tag boundaries and search within them.
	tagStart := strings.LastIndex(htmlContent[:idx], "<")
	if tagStart == -1 {
		tagStart = 0
	}
	tagEnd := strings.Index(htmlContent[idx:], ">")
	if tagEnd == -1 {
		tagEnd = len(htmlContent) - idx
	}
	tagSlice := htmlContent[tagStart : idx+tagEnd]
	if eq := strings.Index(tagSlice, "content=\""); eq != -1 {
		rest := tagSlice[eq+9:]
		if end := strings.Index(rest, "\""); end != -1 {
			return rest[:end]
		}
	}
	return ""
}

func StripHTML(s string) string {
	var result strings.Builder
	inTag := false
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == '<' {
			// Check if this is a <script> or <style> tag — if so, skip
			// everything until the matching closing tag to avoid leaking
			// JS/CSS content (which may contain API keys, comments, etc.)
			// into the extracted text.
			// R58-B3: lowercase the substring (not the full string) —
			// strings.ToLower can change byte length for some Unicode
			// chars (İ 2→1B, ẞ 3→2B, K Kelvin 3→1B), causing the pre-
			// computed lower copy to misalign with s's byte offsets.
			rest := s[i:]
			restLower := strings.ToLower(rest)
			if strings.HasPrefix(restLower, "<script") {
				if endIdx := strings.Index(restLower, "</script>"); endIdx != -1 {
					i += endIdx + len("</script>")
					continue
				}
			} else if strings.HasPrefix(restLower, "<style") {
				if endIdx := strings.Index(restLower, "</style>"); endIdx != -1 {
					i += endIdx + len("</style>")
					continue
				}
			}
			inTag = true
			i += size
			continue
		}
		if r == '>' {
			inTag = false
			result.WriteString(" ")
			i += size
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
		i += size
	}

	text := result.String()
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", " ")

	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// R50-B3: decode HTML entities (&amp; → &, &lt; → <, &#39; → ', &nbsp; → space)
	// so extracted text is clean for AI processing and keyword matching.
	text = html.UnescapeString(text)

	return strings.TrimSpace(text)
}

// emailRegex is compiled once at package initialization; IsValidEmail is called
// on every auth request, so recompiling the regex each call wastes CPU.
// R44-B2: Tightened domain part to reject consecutive dots, leading/trailing
// dots, and leading hyphens in domain labels. Added 254-char length check
// per RFC 5321.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@([a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// validLangs is the single source of truth for supported language codes.
// Used by handlers to validate the "lang" query/body parameter before
// passing it to AI prompts, cache keys, or persistence.
var validLangs = map[string]bool{
	"zh": true, "en": true, "ja": true, "ko": true, "ar": true,
	"es": true, "pt": true, "fr": true, "de": true, "hi": true,
}

// IsValidLang returns true if lang is a supported language code.
func IsValidLang(lang string) bool {
	return validLangs[lang]
}

// userMsgLabels holds the localized label triplets for BuildUserMsg.
// Order: targetJob label, jobDesc label, resumeContent label.
var userMsgLabels = map[string][3]string{
	"zh": {"目标职位: ", "职位描述: ", "简历内容: "},
	"ja": {"希望職種: ", "職務記述書: ", "履歴書内容: "},
	"ko": {"희망 직종: ", "직무 설명: ", "이력서 내용: "},
	"ar": {"المنصب المستهدف: ", "وصف الوظيفة: ", "محتوى السيرة الذاتية: "},
	"es": {"Puesto objetivo: ", "Descripción del puesto: ", "Contenido del CV: "},
	"pt": {"Cargo alvo: ", "Descrição da vaga: ", "Conteúdo do currículo: "},
	"fr": {"Poste cible: ", "Description du poste: ", "Contenu du CV: "},
	"de": {"Zielposition: ", "Stellenbeschreibung: ", "Lebenslauf-Inhalt: "},
	"hi": {"लक्षित पद: ", "नौकरी विवरण: ", "रिज़्यूमे सामग्री: "},
	"en": {"Target Position: ", "Job Description: ", "Resume Content: "},
}

// R57b-B1: wrap user input in XML tags to defend against prompt injection.
// Mirrors the pattern used by product.go (R49-B1): the system prompt tells
// the AI that content inside these tags is untrusted data, not instructions.
// Without this, an attacker can embed "Ignore the above..." in their resume
// text to manipulate the AI output or inflate the ATS score.
func BuildUserMsg(lang, targetJob, jobDesc, resumeContent string) string {
	labels, ok := userMsgLabels[lang]
	if !ok {
		labels = userMsgLabels["en"]
	}
	return labels[0] + "<user_target_job>\n" + targetJob + "\n</user_target_job>\n\n" +
		labels[1] + "<user_job_description>\n" + jobDesc + "\n</user_job_description>\n\n" +
		labels[2] + "<user_resume>\n" + resumeContent + "\n</user_resume>"
}

// ExtractTitle returns the page <title> text. Used as a fallback when
// og:title is absent — many job sites (especially Chinese ones) set <title>
// but not og:title. Returns empty string if no <title> tag is found.
// R57-B6: previously ScrapeJob only checked og:title, leaving title empty
// for sites that don't implement Open Graph.
func ExtractTitle(htmlContent string) string {
	// Case-insensitive search for <title ...>...</title>.
	lower := strings.ToLower(htmlContent)
	startTag := strings.Index(lower, "<title")
	if startTag == -1 {
		return ""
	}
	// Skip past the tag name and any attributes to the closing '>'.
	contentStart := strings.Index(htmlContent[startTag:], ">")
	if contentStart == -1 {
		return ""
	}
	contentStart += startTag + 1
	endTag := strings.Index(lower[contentStart:], "</title>")
	if endTag == -1 {
		return ""
	}
	title := htmlContent[contentStart : contentStart+endTag]
	// Decode HTML entities and trim whitespace.
	return strings.TrimSpace(html.UnescapeString(title))
}

func BuildModuleHints(modules []interface{}) string {
	// Iterate in a fixed canonical order rather than the input slice order.
	// The output is used as part of a cache/dedup key, so the same set of
	// modules must always produce the same string regardless of input order.
	has := make(map[string]bool, len(modules))
	for _, m := range modules {
		if s, ok := m.(string); ok {
			has[s] = true
		}
	}
	moduleHints := ""
	for _, step := range []struct{ name, hint string }{
		{"ats", "1. Extract and match ATS keywords from the job description\n"},
		{"star", "2. Rewrite work experience using STAR method (Situation-Task-Action-Result)\n"},
		{"quant", "3. Add quantified achievements with specific metrics and data\n"},
		{"summary", "4. Optimize professional summary to highlight core competencies\n"},
		{"format", "5. Optimize resume structure, formatting, and layout\n"},
	} {
		if has[step.name] {
			moduleHints += step.hint
		}
	}
	return moduleHints
}
