package url

import (
	"os"
	"strings"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

// GetBaseURL returns the base URL for the application
func GetBaseURL() string {
	// Check for environment variable first
	if baseURL := os.Getenv(constants.EnvBaseURL); baseURL != "" {
		return strings.TrimSuffix(baseURL, "/")
	}

	// Default to localhost for development
	return constants.DefaultBaseURL
}
