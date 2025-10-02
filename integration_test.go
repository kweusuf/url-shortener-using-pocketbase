package main

import (
	"strings"
	"testing"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/auth"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/generator"
	"github.com/kweusuf/pocketbase-demo/pkg/utils/url"
)

// TestAuthFunctions tests the core authentication functions
func TestAuthFunctions(t *testing.T) {
	t.Log("Testing authentication functions...")

	// Test email validation
	validEmail := "test@example.com"
	invalidEmail := "invalid"

	err := auth.ValidateEmail(validEmail)
	if err != nil {
		t.Errorf("Expected valid email, got error: %v", err)
	}

	err = auth.ValidateEmail(invalidEmail)
	if err == nil {
		t.Error("Expected error for invalid email")
	}

	// Test password validation
	validPassword := "password123"
	shortPassword := "123"

	err = auth.ValidatePassword(validPassword)
	if err != nil {
		t.Errorf("Expected valid password, got error: %v", err)
	}

	err = auth.ValidatePassword(shortPassword)
	if err == nil {
		t.Error("Expected error for short password")
	}

	// Test password hashing
	hash, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Errorf("Hashing failed: %v", err)
	}
	if hash == validPassword {
		t.Error("Hash should not equal original password")
	}

	// Test password verification
	err = auth.VerifyPassword(validPassword, hash)
	if err != nil {
		t.Errorf("Password verification failed: %v", err)
	}

	err = auth.VerifyPassword("wrongpassword", hash)
	if err == nil {
		t.Error("Incorrect password should fail verification")
	}

	t.Log("Auth functions tests passed")
}

// TestGeneratorFunctions tests the short code generation
func TestGeneratorFunctions(t *testing.T) {
	t.Log("Testing generator functions...")

	// Test short code generation
	code1 := generator.GenerateShortCode()
	code2 := generator.GenerateShortCode()

	if len(code1) == 0 {
		t.Error("Generated short code should not be empty")
	}

	if len(code1) != constants.ShortCodeLength {
		t.Errorf("Expected code length %d, got %d", constants.ShortCodeLength, len(code1))
	}

	if code1 == code2 {
		t.Log("Warning: Generated codes are identical (this is rare but possible)")
	}

	t.Log("Generator functions tests passed")
}

// TestURLFunctions tests URL utility functions
func TestURLFunctions(t *testing.T) {
	t.Log("Testing URL functions...")

	baseURL := url.GetBaseURL()
	if baseURL == "" {
		t.Error("Base URL should not be empty")
	}

	// Should contain protocol
	if !strings.Contains(baseURL, "://") {
		t.Error("Base URL should contain protocol")
	}

	t.Log("URL functions tests passed")
}

// Integration test demonstrating the flow (skipped due to complexity)
func TestRegistrationLoginShortenFlow(t *testing.T) {
	t.Skip("Full integration test skipped due to PocketBase initialization complexity")

	// This test documents the expected flow for the integration:
	// 1. Register user via POST /api/auth/register
	//    - Validates email/password
	//    - Creates user in PocketBase auth collection
	//    - Returns JWT token

	// 2. Login user via POST /api/auth/login
	//    - Validates credentials
	//    - Returns JWT token

	// 3. Shorten URL via POST /api/shorten
	//    - Validates URL format
	//    - Generates short code
	//    - Stores URL in database with user association
	//    - Returns short code and short URL

	// 4. Access shortened URL via GET /{shortCode}
	//    - Retrieves URL from database
	//    - Increments click count
	//    - Redirects to original URL
	//    - Broadcasts click update via WebSocket

	// 5. Check stats via GET /api/stats/{shortCode}
	//    - Returns click count and other statistics

	// 6. Get recent URLs via authenticated GET /api/recent
	//    - Returns user's recent shortened URLs

	t.Logf(`Integration test flow documented:
	1. User registration (auth.RegisterAuthRoutes)
	2. User login (auth.RegisterAuthRoutes)
	3. URL shortening (transport/http/routes.RegisterHTTPRoutes)
	4. URL access and redirect (transport/http/routes.RegisterHTTPRoutes)
	5. Statistics retrieval (transport/http/routes.RegisterHTTPRoutes)
	6. Recent URLs listing (transport/http/routes.RegisterHTTPRoutes)
	`)

	// Test individual components that don't require full PocketBase setup
	TestAuthFunctions(t)
	TestGeneratorFunctions(t)
	TestURLFunctions(t)
}
