package generator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

func TestGenerateShortCode(t *testing.T) {
	// Test basic functionality
	code1 := GenerateShortCode()
	code2 := GenerateShortCode()

	// Verify codes are generated
	if code1 == "" {
		t.Error("Generated short code should not be empty")
	}

	if code2 == "" {
		t.Error("Generated short code should not be empty")
	}

	// Verify codes are unique
	if code1 == code2 {
		t.Error("Generated short codes should be unique")
	}

	// Verify code length
	if len(code1) != constants.ShortCodeLength {
		t.Errorf("Expected short code length of %d, got %d", constants.ShortCodeLength, len(code1))
	}

	if len(code2) != constants.ShortCodeLength {
		t.Errorf("Expected short code length of %d, got %d", constants.ShortCodeLength, len(code2))
	}

	// Verify all characters are valid base62 characters
	for _, char := range code1 {
		if !strings.Contains(constants.Base62Charset, string(char)) {
			t.Errorf("Short code contains invalid character: %c", char)
		}
	}

	for _, char := range code2 {
		if !strings.Contains(constants.Base62Charset, string(char)) {
			t.Errorf("Short code contains invalid character: %c", char)
		}
	}
}

func TestGenerateShortCodeUniqueness(t *testing.T) {
	// Test multiple generations for uniqueness
	codes := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		code := GenerateShortCode()
		if codes[code] {
			t.Errorf("Duplicate short code generated at iteration %d: %s", i, code)
		}
		codes[code] = true

		// Verify length each time
		if len(code) != constants.ShortCodeLength {
			t.Errorf("Short code %d has incorrect length: %d", i, len(code))
		}
	}

	// Verify we generated the expected number of unique codes
	if len(codes) != iterations {
		t.Errorf("Expected %d unique short codes, got %d", iterations, len(codes))
	}
}

func TestGenerateShortCodeConsistency(t *testing.T) {
	// Test that the function is consistent in its behavior
	for i := 0; i < 100; i++ {
		code := GenerateShortCode()

		// Should always be exactly 6 characters
		if len(code) != constants.ShortCodeLength {
			t.Errorf("Short code %d has incorrect length: %d", i, len(code))
		}

		// Should always use valid base62 characters
		for _, char := range code {
			if !strings.Contains(constants.Base62Charset, string(char)) {
				t.Errorf("Short code %d contains invalid character: %c", i, char)
			}
		}

		// Should not be empty
		if code == "" {
			t.Errorf("Short code %d is empty", i)
		}

		// Should not contain only zeros (edge case)
		if code == "000000" {
			t.Logf("Warning: Generated short code with all zeros: %s", code)
		}
	}
}

func TestGenerateShortCodeDistribution(t *testing.T) {
	// Test that the generated codes have good distribution
	// Generate many codes and check character distribution
	codes := make(map[string]bool)
	charCount := make(map[rune]int)

	iterations := 10000
	for i := 0; i < iterations; i++ {
		code := GenerateShortCode()
		codes[code] = true

		for _, char := range code {
			charCount[char]++
		}
	}

	// Verify we got a good number of unique codes
	uniqueCodes := len(codes)
	expectedMin := iterations * 90 / 100 // At least 90% unique
	if uniqueCodes < expectedMin {
		t.Errorf("Expected at least %d unique codes, got %d", expectedMin, uniqueCodes)
	}

	// Check that all base62 characters are being used
	for _, char := range constants.Base62Charset {
		if charCount[char] == 0 {
			t.Errorf("Character %c was never used in generated codes", char)
		}
	}

	// Check that no character is used excessively (basic distribution test)
	totalChars := 0
	for _, count := range charCount {
		totalChars += count
	}

	expectedPerChar := totalChars / len(constants.Base62Charset)
	tolerance := expectedPerChar / 2 // 50% tolerance

	for char, count := range charCount {
		if count < expectedPerChar-tolerance || count > expectedPerChar+tolerance {
			t.Logf("Character %c has uneven distribution: %d (expected around %d)", char, count, expectedPerChar)
		}
	}
}

func TestEncodeBase62(t *testing.T) {
	// Test the base62 encoding function directly
	testCases := []struct {
		input    uint32
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "A"},
		{35, "Z"},
		{36, "a"},
		{61, "z"},
		{62, "10"},
		{100, "1c"},
		{1000, "G8"},
		{10000, "2bI"},
		{100000, "Q0u"},
		{1000000, "4C92"},
	}

	for _, tc := range testCases {
		result := encodeBase62(tc.input)
		if result != tc.expected {
			t.Errorf("encodeBase62(%d) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestEncodeBase62EdgeCases(t *testing.T) {
	// Test edge cases for base62 encoding
	testCases := []struct {
		name     string
		input    uint32
		validate func(t *testing.T, result string)
	}{
		{
			name:  "zero",
			input: 0,
			validate: func(t *testing.T, result string) {
				if result != "0" {
					t.Errorf("Expected '0', got %s", result)
				}
			},
		},
		{
			name:  "max uint32",
			input: ^uint32(0), // max uint32 value
			validate: func(t *testing.T, result string) {
				if result == "" {
					t.Error("Result should not be empty for max uint32")
				}
				if len(result) == 0 {
					t.Error("Result length should be > 0 for max uint32")
				}
			},
		},
		{
			name:  "power of 62",
			input: 62 * 62, // 3844
			validate: func(t *testing.T, result string) {
				if result == "" {
					t.Error("Result should not be empty")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := encodeBase62(tc.input)
			tc.validate(t, result)

			// Verify all characters are valid base62
			for _, char := range result {
				if !strings.Contains(constants.Base62Charset, string(char)) {
					t.Errorf("Result contains invalid character: %c", char)
				}
			}
		})
	}
}

func TestGenerateShortCodeLength(t *testing.T) {
	// Test that all generated codes have exactly the expected length
	for i := 0; i < 1000; i++ {
		code := GenerateShortCode()
		if len(code) != constants.ShortCodeLength {
			t.Errorf("Short code %d has length %d, expected %d", i, len(code), constants.ShortCodeLength)
		}
	}
}

func TestGenerateShortCodeCharacterSet(t *testing.T) {
	// Test that all generated codes only use valid base62 characters
	validChars := make(map[rune]bool)
	for _, char := range constants.Base62Charset {
		validChars[char] = true
	}

	for i := 0; i < 1000; i++ {
		code := GenerateShortCode()
		for _, char := range code {
			if !validChars[char] {
				t.Errorf("Short code %d contains invalid character: %c", i, char)
			}
		}
	}
}

func BenchmarkGenerateShortCode(b *testing.B) {
	// Benchmark short code generation performance
	for i := 0; i < b.N; i++ {
		GenerateShortCode()
	}
}

func BenchmarkGenerateShortCodeParallel(b *testing.B) {
	// Benchmark parallel short code generation
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateShortCode()
		}
	})
}

func BenchmarkEncodeBase62(b *testing.B) {
	// Benchmark base62 encoding performance
	testValues := []uint32{0, 1, 100, 1000, 10000, 100000, 1000000, 10000000}

	for _, value := range testValues {
		b.Run(fmt.Sprintf("value_%d", value), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				encodeBase62(value)
			}
		})
	}
}
