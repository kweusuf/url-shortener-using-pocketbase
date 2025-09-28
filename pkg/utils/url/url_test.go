package url

import (
	"os"
	"testing"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

func TestGetBaseURL(t *testing.T) {
	// Test default behavior (no environment variable)
	url := GetBaseURL()
	if url != constants.DefaultBaseURL {
		t.Errorf("Expected default base URL %s, got %s", constants.DefaultBaseURL, url)
	}
}

func TestGetBaseURLWithEnvironmentVariable(t *testing.T) {
	// Set environment variable
	testURL := "https://test.example.com:3000"
	os.Setenv(constants.EnvBaseURL, testURL)

	url := GetBaseURL()
	if url != testURL {
		t.Errorf("Expected base URL %s, got %s", testURL, url)
	}

	// Clean up
	os.Unsetenv(constants.EnvBaseURL)
}

func TestGetBaseURLWithTrailingSlash(t *testing.T) {
	// Set environment variable with trailing slash
	testURL := "https://test.example.com:3000/"
	os.Setenv(constants.EnvBaseURL, testURL)

	url := GetBaseURL()
	expectedURL := "https://test.example.com:3000" // Should remove trailing slash
	if url != expectedURL {
		t.Errorf("Expected base URL %s, got %s", expectedURL, url)
	}

	// Clean up
	os.Unsetenv(constants.EnvBaseURL)
}

func TestGetBaseURLWithMultipleTrailingSlashes(t *testing.T) {
	// Set environment variable with multiple trailing slashes
	testURL := "https://test.example.com:3000//"
	os.Setenv(constants.EnvBaseURL, testURL)

	url := GetBaseURL()
	expectedURL := "https://test.example.com:3000/" // Should remove only one trailing slash
	if url != expectedURL {
		t.Errorf("Expected base URL %s, got %s", expectedURL, url)
	}

	// Clean up
	os.Unsetenv(constants.EnvBaseURL)
}

func TestGetBaseURLWithDifferentProtocols(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"http://localhost:8080", "http://localhost:8080"},
		{"https://example.com", "https://example.com"},
		{"http://192.168.1.1:3000", "http://192.168.1.1:3000"},
		{"https://myapp.herokuapp.com", "https://myapp.herokuapp.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			os.Setenv(constants.EnvBaseURL, tc.input)
			defer os.Unsetenv(constants.EnvBaseURL)

			url := GetBaseURL()
			if url != tc.expected {
				t.Errorf("Expected base URL %s, got %s", tc.expected, url)
			}
		})
	}
}

func TestGetBaseURLWithPorts(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"localhost:3000", "localhost:3000"},
		{"example.com:8080", "example.com:8080"},
		{"192.168.1.1:9000", "192.168.1.1:9000"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			os.Setenv(constants.EnvBaseURL, tc.input)
			defer os.Unsetenv(constants.EnvBaseURL)

			url := GetBaseURL()
			if url != tc.expected {
				t.Errorf("Expected base URL %s, got %s", tc.expected, url)
			}
		})
	}
}

func TestGetBaseURLWithPaths(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"https://example.com/api", "https://example.com/api"},
		{"https://example.com/v1", "https://example.com/v1"},
		{"localhost:3000/app", "localhost:3000/app"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			os.Setenv(constants.EnvBaseURL, tc.input)
			defer os.Unsetenv(constants.EnvBaseURL)

			url := GetBaseURL()
			if url != tc.expected {
				t.Errorf("Expected base URL %s, got %s", tc.expected, url)
			}
		})
	}
}

func TestGetBaseURLConsistency(t *testing.T) {
	// Test that multiple calls return the same result
	testURL := "https://consistent-test.com"
	os.Setenv(constants.EnvBaseURL, testURL)
	defer os.Unsetenv(constants.EnvBaseURL)

	// Call multiple times
	for i := 0; i < 100; i++ {
		url := GetBaseURL()
		if url != testURL {
			t.Errorf("Inconsistent result on call %d: expected %s, got %s", i, testURL, url)
		}
	}
}

func TestGetBaseURLThreadSafety(t *testing.T) {
	// Test thread safety by calling from multiple goroutines
	testURL := "https://thread-safety-test.com"
	os.Setenv(constants.EnvBaseURL, testURL)
	defer os.Unsetenv(constants.EnvBaseURL)

	done := make(chan bool, 10)

	// Start multiple goroutines
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < 100; j++ {
				url := GetBaseURL()
				if url != testURL {
					t.Errorf("Thread safety test failed: expected %s, got %s", testURL, url)
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetBaseURLEnvironmentVariablePriority(t *testing.T) {
	// Test that environment variable takes priority over default

	// First, ensure no environment variable is set
	os.Unsetenv(constants.EnvBaseURL)
	url1 := GetBaseURL()
	if url1 != constants.DefaultBaseURL {
		t.Errorf("Expected default base URL %s, got %s", constants.DefaultBaseURL, url1)
	}

	// Set environment variable
	testURL := "https://env-priority-test.com"
	os.Setenv(constants.EnvBaseURL, testURL)
	url2 := GetBaseURL()
	if url2 != testURL {
		t.Errorf("Expected environment base URL %s, got %s", testURL, url2)
	}

	// Clean up
	os.Unsetenv(constants.EnvBaseURL)
}

func TestGetBaseURLWithEmptyEnvironmentVariable(t *testing.T) {
	// Test with empty environment variable
	os.Setenv(constants.EnvBaseURL, "")
	defer os.Unsetenv(constants.EnvBaseURL)

	url := GetBaseURL()
	if url != constants.DefaultBaseURL {
		t.Errorf("Expected default base URL when env var is empty, got %s", url)
	}
}

func TestGetBaseURLWithWhitespaceEnvironmentVariable(t *testing.T) {
	// Test with whitespace-only environment variable
	os.Setenv(constants.EnvBaseURL, "   ")
	defer os.Unsetenv(constants.EnvBaseURL)

	url := GetBaseURL()
	if url != "   " {
		t.Errorf("Expected whitespace URL when env var is whitespace, got %s", url)
	}
}

func TestGetBaseURLRealWorldExamples(t *testing.T) {
	examples := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Production HTTPS",
			input:    "https://api.myservice.com",
			expected: "https://api.myservice.com",
		},
		{
			name:     "Development with port",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "Docker container",
			input:    "http://0.0.0.0:3000",
			expected: "http://0.0.0.0:3000",
		},
		{
			name:     "Heroku deployment",
			input:    "https://myapp.herokuapp.com",
			expected: "https://myapp.herokuapp.com",
		},
		{
			name:     "AWS load balancer",
			input:    "https://myapp-123456789.us-east-1.elb.amazonaws.com",
			expected: "https://myapp-123456789.us-east-1.elb.amazonaws.com",
		},
		{
			name:     "Kubernetes service",
			input:    "http://myapp-service:8080",
			expected: "http://myapp-service:8080",
		},
	}

	for _, example := range examples {
		t.Run(example.name, func(t *testing.T) {
			os.Setenv(constants.EnvBaseURL, example.input)
			defer os.Unsetenv(constants.EnvBaseURL)

			url := GetBaseURL()
			if url != example.expected {
				t.Errorf("Expected %s, got %s", example.expected, url)
			}
		})
	}
}

func BenchmarkGetBaseURL(b *testing.B) {
	// Benchmark GetBaseURL performance
	testURL := "https://benchmark-test.com"
	os.Setenv(constants.EnvBaseURL, testURL)
	defer os.Unsetenv(constants.EnvBaseURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetBaseURL()
	}
}

func BenchmarkGetBaseURLWithoutEnv(b *testing.B) {
	// Benchmark GetBaseURL performance without environment variable
	os.Unsetenv(constants.EnvBaseURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetBaseURL()
	}
}
