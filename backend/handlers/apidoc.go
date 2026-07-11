package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type APIInfo struct {
	Version  string                 `json:"version"`
	Title    string                 `json:"title"`
	BasePath string                 `json:"base_path"`
	Paths    map[string]PathInfo    `json:"paths"`
}

type PathInfo struct {
	Get  *EndpointInfo `json:"get,omitempty"`
	Post *EndpointInfo `json:"post,omitempty"`
}

type EndpointInfo struct {
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Auth        bool     `json:"auth"`
}

func GetAPIDoc(c *fiber.Ctx) error {
	api := APIInfo{
		Version:  "2.1.0",
		Title:    "ResumeTake API",
		BasePath: "/api/v1",
		Paths: map[string]PathInfo{
			"/health": {
				Get: &EndpointInfo{
					Summary:     "Health check",
					Description: "Check API health status",
					Tags:        []string{"system"},
				},
			},
			"/auth/send-code": {
				Post: &EndpointInfo{
					Summary:     "Send verification code",
					Description: "Send email verification code for registration",
					Tags:        []string{"auth"},
				},
			},
			"/auth/verify-code": {
				Post: &EndpointInfo{
					Summary:     "Verify code",
					Description: "Verify email verification code",
					Tags:        []string{"auth"},
				},
			},
			"/auth/register": {
				Post: &EndpointInfo{
					Summary:     "Register",
					Description: "Register new user account",
					Tags:        []string{"auth"},
				},
			},
			"/auth/login": {
				Post: &EndpointInfo{
					Summary:     "Login",
					Description: "Login with email and password",
					Tags:        []string{"auth"},
				},
			},
			"/auth/me": {
				Get: &EndpointInfo{
					Summary:     "Get current user",
					Description: "Get current authenticated user info",
					Tags:        []string{"auth"},
					Auth:        true,
				},
			},
			"/resumes": {
				Post: &EndpointInfo{
					Summary:     "Create resume",
					Description: "Create a new resume",
					Tags:        []string{"resume"},
				},
			},
			"/resumes/:id": {
				Get: &EndpointInfo{
					Summary:     "Get resume",
					Description: "Get resume by ID",
					Tags:        []string{"resume"},
				},
			},
			"/upload": {
				Post: &EndpointInfo{
					Summary:     "Upload file",
					Description: "Upload PDF/TXT/MD file for text extraction",
					Tags:        []string{"resume"},
					Auth:        true,
				},
			},
			"/optimize": {
				Post: &EndpointInfo{
					Summary:     "Optimize resume",
					Description: "AI-powered resume optimization",
					Tags:        []string{"ai"},
					Auth:        true,
				},
			},
			"/optimize-stream": {
				Post: &EndpointInfo{
					Summary:     "Stream optimization",
					Description: "SSE streaming AI resume optimization",
					Tags:        []string{"ai"},
					Auth:        true,
				},
			},
			"/perspective": {
				Post: &EndpointInfo{
					Summary:     "Analyze perspective",
					Description: "Four-perspective resume analysis",
					Tags:        []string{"ai"},
					Auth:        true,
				},
			},
			"/scrape-job": {
				Post: &EndpointInfo{
					Summary:     "Scrape job",
					Description: "Scrape job description from URL",
					Tags:        []string{"job"},
				},
			},
			"/pricing": {
				Get: &EndpointInfo{
					Summary:     "Get pricing",
					Description: "Get pricing tiers",
					Tags:        []string{"payment"},
				},
			},
			"/templates": {
				Get: &EndpointInfo{
					Summary:     "Get templates",
					Description: "Get available resume templates",
					Tags:        []string{"resume"},
				},
			},
			"/create-paypal-order": {
				Post: &EndpointInfo{
					Summary:     "Create PayPal order",
					Description: "Create PayPal payment order",
					Tags:        []string{"payment"},
					Auth:        true,
				},
			},
			"/capture-paypal-order": {
				Post: &EndpointInfo{
					Summary:     "Capture PayPal order",
					Description: "Capture PayPal payment",
					Tags:        []string{"payment"},
					Auth:        true,
				},
			},
			"/metrics": {
				Get: &EndpointInfo{
					Summary:     "Metrics",
					Description: "Performance metrics",
					Tags:        []string{"system"},
				},
			},
		},
	}

	return c.JSON(api)
}
