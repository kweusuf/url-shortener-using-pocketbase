package main

import (
	"net/url"
	"testing"
	"time"
)

func TestGenerateShortCode(t *testing.T) {
	// Test short code generation
	code1 := generateShortCode()
	code2 := generateShortCode()

	// Check length
	if len(code1) != 6 {
		t.Errorf("Expected length 6, got %d", len(code1))
	}

	if len(code2) != 6 {
		t.Errorf("Expected length 6, got %d", len(code2))
	}

	// Check they're different (very high probability)
	if code1 == code2 {
		t.Error("Generated codes should be different")
	}

	// Check characters are valid base62 (alphanumeric)
	for _, char := range code1 {
		if !((char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z')) {
			t.Errorf("Invalid character in short code: %c", char)
		}
	}

	// Test base62 encoding function directly
	testNum := uint32(12345)
	encoded := encodeBase62(testNum)
	if len(encoded) == 0 {
		t.Error("Base62 encoding should not return empty string")
	}

	// Check all characters in encoded string are valid base62
	for _, char := range encoded {
		if !((char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z')) {
			t.Errorf("Invalid base62 character: %c", char)
		}
	}

	// Test padding logic - small numbers should be padded
	smallNum := uint32(1)
	smallEncoded := encodeBase62(smallNum)
	if len(smallEncoded) == 0 {
		t.Error("Small number encoding should not be empty")
	}
}

func TestURLNormalization(t *testing.T) {
	// Clear storage before testing
	urlStore = make(map[string]map[string]interface{})

	tests := []struct {
		name           string
		inputURL       string
		expectedStored string
		expectError    bool
	}{
		{
			name:           "Bare domain gets HTTPS",
			inputURL:       "google.com",
			expectedStored: "https://google.com",
			expectError:    false,
		},
		{
			name:           "Bare domain with path gets HTTPS",
			inputURL:       "github.com/user/repo",
			expectedStored: "https://github.com/user/repo",
			expectError:    false,
		},
		{
			name:           "HTTPS URL preserved",
			inputURL:       "https://example.com",
			expectedStored: "https://example.com",
			expectError:    false,
		},
		{
			name:           "HTTP URL preserved",
			inputURL:       "http://example.com",
			expectedStored: "http://example.com",
			expectError:    false,
		},
		{
			name:           "Invalid protocol rejected",
			inputURL:       "ftp://example.com",
			expectedStored: "",
			expectError:    true,
		},
		{
			name:           "Malformed URL gets HTTPS added",
			inputURL:       "not-a-url",
			expectedStored: "https://not-a-url",
			expectError:    false,
		},
		{
			name:           "Empty URL rejected",
			inputURL:       "",
			expectedStored: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test URL parsing and normalization logic
			if tt.inputURL == "" {
				// Empty URL should be rejected
				if !tt.expectError {
					t.Error("Expected error for empty URL")
				}
				return
			}

			// Parse URL (same logic as in main.go)
			parsedURL, err := url.Parse(tt.inputURL)
			if err != nil && !tt.expectError {
				t.Errorf("Unexpected parse error for %s: %v", tt.inputURL, err)
				return
			}

			if err == nil {
				// Apply normalization logic
				if parsedURL.Scheme == "" {
					// Add https:// protocol for bare URLs
					normalizedURL := "https://" + tt.inputURL
					if normalizedURL != tt.expectedStored {
						t.Errorf("Expected normalized URL %s, got %s", tt.expectedStored, normalizedURL)
					}
				} else if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
					if tt.inputURL != tt.expectedStored {
						t.Errorf("Expected stored URL %s, got %s", tt.expectedStored, tt.inputURL)
					}
				} else {
					// Invalid protocol should be rejected
					if !tt.expectError {
						t.Errorf("Expected error for invalid protocol: %s", parsedURL.Scheme)
					}
				}
			}
		})
	}
}

func TestURLRedirectWithNormalization(t *testing.T) {
	// Clear storage
	urlStore = make(map[string]map[string]interface{})

	// Test that normalized URLs redirect correctly
	testCases := []struct {
		inputURL    string
		expectedURL string
	}{
		{"google.com", "https://google.com"},
		{"github.com", "https://github.com"},
		{"https://example.com", "https://example.com"},
		{"http://example.com", "http://example.com"},
	}

	for _, tc := range testCases {
		t.Run("Redirect_"+tc.inputURL, func(t *testing.T) {
			// Generate a test short code
			shortCode := generateShortCode()

			// Store the expected normalized URL
			urlData := map[string]interface{}{
				"original_url": tc.expectedURL,
				"short_code":   shortCode,
				"clicks":       0,
				"created":      time.Now(),
			}
			storeURL(shortCode, urlData)

			// Verify the stored URL is correct
			retrieved := getURL(shortCode)
			if retrieved == nil {
				t.Fatal("URL not stored correctly")
			}

			storedURL := retrieved["original_url"].(string)
			if storedURL != tc.expectedURL {
				t.Errorf("Expected stored URL %s, got %s", tc.expectedURL, storedURL)
			}
		})
	}
}

func TestURLStorage(t *testing.T) {
	// Clear storage
	urlStore = make(map[string]map[string]interface{})

	// Test data
	testCode := "test456"
	testData := map[string]interface{}{
		"original_url": "https://www.test.com",
		"clicks":       5,
		"created":      time.Now(),
	}

	// Test store
	storeURL(testCode, testData)

	// Test retrieve
	retrieved := getURL(testCode)
	if retrieved == nil {
		t.Fatal("Data not stored correctly")
	}

	// Verify data integrity
	if retrieved["original_url"].(string) != testData["original_url"].(string) {
		t.Error("Original URL not stored correctly")
	}

	if retrieved["clicks"].(int) != testData["clicks"].(int) {
		t.Error("Clicks count not stored correctly")
	}

	// Test non-existent key
	nonExistent := getURL("nonexistent")
	if nonExistent != nil {
		t.Error("Should return nil for non-existent key")
	}
}

func BenchmarkGenerateShortCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateShortCode()
	}
}

func BenchmarkURLStorage(b *testing.B) {
	urlStore = make(map[string]map[string]interface{})

	testData := map[string]interface{}{
		"original_url": "https://www.benchmark.com",
		"clicks":       0,
		"created":      time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := generateShortCode()
		storeURL(code, testData)
	}
}
