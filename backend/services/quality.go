package services

import (
	"encoding/json"
	"strings"
)

type QualityScore struct {
	IsValid       bool     `json:"is_valid"`
	JSONValid     bool     `json:"json_valid"`
	HasSummary    bool     `json:"has_summary"`
	HasSkills     bool     `json:"has_skills"`
	HasExperience bool     `json:"has_experience"`
	ATSScore      float64  `json:"ats_score"`
	OverallScore  float64  `json:"overall_score"`
	Issues        []string `json:"issues"`
}

func ValidateAIResponse(response string) QualityScore {
	score := QualityScore{
		Issues: []string{},
	}

	response = strings.TrimSpace(response)

	if len(response) == 0 {
		score.Issues = append(score.Issues, "empty_response")
		return score
	}

	// Strip markdown code fences case-insensitively — some AI providers
	// return ```JSON (uppercase) which the previous case-sensitive check
	// missed, causing valid responses to be rejected as invalid_json.
	lowerResp := strings.ToLower(response)
	if strings.HasPrefix(lowerResp, "```json") {
		response = response[len("```json"):]
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(lowerResp, "```") {
		response = response[3:]
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		score.Issues = append(score.Issues, "invalid_json")
		score.OverallScore = 0
		return score
	}
	score.JSONValid = true

	if optimized, ok := result["optimized_content"].(map[string]interface{}); ok {
		if s, ok := optimized["summary"].(string); ok && s != "" {
			score.HasSummary = true
		}
		if skills, ok := optimized["skills"].([]interface{}); ok && len(skills) > 0 {
			score.HasSkills = true
		}
		if exp, ok := optimized["experience"].([]interface{}); ok && len(exp) > 0 {
			score.HasExperience = true
		}
	}

	if atsScore, ok := result["ats_score"].(float64); ok {
		// R53-B3: clamp to valid range — AI may return out-of-range values
		// (e.g. 9999 or negative), which break frontend display and logic.
		if atsScore < 0 {
			atsScore = 0
		} else if atsScore > 100 {
			atsScore = 100
		}
		score.ATSScore = atsScore
		if atsScore < 50 {
			score.Issues = append(score.Issues, "low_ats_score")
		}
	}

	score.OverallScore = 0
	if score.JSONValid {
		score.OverallScore += 40
	}
	if score.HasSummary {
		score.OverallScore += 20
	}
	if score.HasSkills {
		score.OverallScore += 20
	}
	if score.HasExperience {
		score.OverallScore += 20
	}

	score.IsValid = score.OverallScore >= 60 && score.JSONValid
	return score
}
