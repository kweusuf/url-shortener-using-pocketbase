package generator

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

// GeneratePocketBaseID generates a PocketBase-style ID
func GeneratePocketBaseID() string {
	bytes := make([]byte, 7)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return constants.PBIDPrefix + fmt.Sprintf("%07x", time.Now().UnixNano()%0x10000000)
	}

	// Convert to hex and lowercase
	hexStr := fmt.Sprintf("%014x", bytes)
	return constants.PBIDPrefix + strings.ToLower(hexStr)
}
