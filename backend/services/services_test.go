package services

import (
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<div><span>Test</span></div>", "Test"},
		{"No HTML", "No HTML"},
		{"<br>", ""},
		{"<p>Line 1</p><p>Line 2</p>", "Line 1 Line 2"},
	}

	for _, test := range tests {
		result := StripHTML(test.input)
		if result != test.expected {
			t.Errorf("StripHTML(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestExtractMeta(t *testing.T) {
	tests := []struct {
		html     string
		name     string
		expected string
	}{
		{`<meta name="description" content="Test description">`, "description", "Test description"},
		{`<meta name="keywords" content="test,keywords">`, "keywords", "test,keywords"},
		{`<html><body>No meta</body></html>`, "description", ""},
	}

	for _, test := range tests {
		result := ExtractMeta(test.html, test.name)
		if result != test.expected {
			t.Errorf("ExtractMeta(%s, %s) = %s, expected %s", test.html, test.name, result, test.expected)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.com", true},
		{"invalid", false},
		{"@domain.com", false},
		{"user@", false},
	}

	for _, test := range tests {
		result := IsValidEmail(test.email)
		if result != test.expected {
			t.Errorf("IsValidEmail(%s) = %v, expected %v", test.email, result, test.expected)
		}
	}
}

func TestGenerateVerificationCode(t *testing.T) {
	code, err := GenerateVerificationCode()
	if err != nil {
		t.Fatalf("GenerateVerificationCode() returned error: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("GenerateVerificationCode() returned code of length %d, expected 6", len(code))
	}
}

func TestBuildUserMsg(t *testing.T) {
	tests := []struct {
		lang         string
		targetJob    string
		jobDesc      string
		resumeContent string
		expected     string
	}{
		// R57b-B1: BuildUserMsg now wraps user input in XML tags for
		// prompt injection defense. Expected strings updated to match.
		{"zh", "Software Engineer", "Job description", "Resume content", "目标职位: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\n职位描述: <user_job_description>\nJob description\n</user_job_description>\n\n简历内容: <user_resume>\nResume content\n</user_resume>"},
		{"ja", "Software Engineer", "Job description", "Resume content", "希望職種: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\n職務記述書: <user_job_description>\nJob description\n</user_job_description>\n\n履歴書内容: <user_resume>\nResume content\n</user_resume>"},
		{"ko", "Software Engineer", "Job description", "Resume content", "희망 직종: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\n직무 설명: <user_job_description>\nJob description\n</user_job_description>\n\n이력서 내용: <user_resume>\nResume content\n</user_resume>"},
		{"ar", "Software Engineer", "Job description", "Resume content", "المنصب المستهدف: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nوصف الوظيفة: <user_job_description>\nJob description\n</user_job_description>\n\nمحتوى السيرة الذاتية: <user_resume>\nResume content\n</user_resume>"},
		{"es", "Software Engineer", "Job description", "Resume content", "Puesto objetivo: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nDescripción del puesto: <user_job_description>\nJob description\n</user_job_description>\n\nContenido del CV: <user_resume>\nResume content\n</user_resume>"},
		{"pt", "Software Engineer", "Job description", "Resume content", "Cargo alvo: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nDescrição da vaga: <user_job_description>\nJob description\n</user_job_description>\n\nConteúdo do currículo: <user_resume>\nResume content\n</user_resume>"},
		{"fr", "Software Engineer", "Job description", "Resume content", "Poste cible: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nDescription du poste: <user_job_description>\nJob description\n</user_job_description>\n\nContenu du CV: <user_resume>\nResume content\n</user_resume>"},
		{"de", "Software Engineer", "Job description", "Resume content", "Zielposition: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nStellenbeschreibung: <user_job_description>\nJob description\n</user_job_description>\n\nLebenslauf-Inhalt: <user_resume>\nResume content\n</user_resume>"},
		{"hi", "Software Engineer", "Job description", "Resume content", "लक्षित पद: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nनौकरी विवरण: <user_job_description>\nJob description\n</user_job_description>\n\nरिज़्यूमे सामग्री: <user_resume>\nResume content\n</user_resume>"},
		{"en", "Software Engineer", "Job description", "Resume content", "Target Position: <user_target_job>\nSoftware Engineer\n</user_target_job>\n\nJob Description: <user_job_description>\nJob description\n</user_job_description>\n\nResume Content: <user_resume>\nResume content\n</user_resume>"},
	}

	for _, test := range tests {
		result := BuildUserMsg(test.lang, test.targetJob, test.jobDesc, test.resumeContent)
		if result != test.expected {
			t.Errorf("BuildUserMsg(%s, %s, %s, %s) = %s, expected %s", test.lang, test.targetJob, test.jobDesc, test.resumeContent, result, test.expected)
		}
	}
}

func TestBuildModuleHints(t *testing.T) {
	tests := []struct {
		modules  []interface{}
		expected string
	}{
		{
			[]interface{}{"ats", "star"},
			"1. Extract and match ATS keywords from the job description\n2. Rewrite work experience using STAR method (Situation-Task-Action-Result)\n",
		},
		{
			[]interface{}{"quant", "summary"},
			"3. Add quantified achievements with specific metrics and data\n4. Optimize professional summary to highlight core competencies\n",
		},
		{
			[]interface{}{},
			"",
		},
	}

	for _, test := range tests {
		result := BuildModuleHints(test.modules)
		if result != test.expected {
			t.Errorf("BuildModuleHints(%v) = %s, expected %s", test.modules, result, test.expected)
		}
	}
}
