package generator

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

// GenerateShortCode generates a random 6-character base62 encoded string
func GenerateShortCode() string {
	const length = 6

	// Generate 4 random bytes (32 bits) which gives us enough entropy
	// 4 bytes = 32 bits = 2^32 possibilities
	// When base62 encoded, this gives us a 6-character string
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based code if random fails
		return fmt.Sprintf("%x", time.Now().UnixNano())[:length]
	}

	// Convert bytes to uint32
	num := uint32(bytes[0])<<24 + uint32(bytes[1])<<16 + uint32(bytes[2])<<8 + uint32(bytes[3])

	// Base62 encode the number
	encoded := encodeBase62(num)

	// Ensure we have at least 6 characters by padding with '0' if necessary
	if len(encoded) < length {
		// Pad with leading zeros
		padding := length - len(encoded)
		padded := strings.Repeat("0", padding) + encoded
		return padded
	}

	// Return first 6 characters if longer
	return encoded[:length]
}

// encodeBase62 encodes a number to base62 string
func encodeBase62(n uint32) string {
	const charset = constants.Base62Charset
	const base = uint32(len(charset))

	if n == 0 {
		return string(charset[0])
	}

	var result []byte
	for n > 0 {
		remainder := n % base
		result = append(result, charset[remainder])
		n = n / base
	}

	// Reverse the result since we built it backwards
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
