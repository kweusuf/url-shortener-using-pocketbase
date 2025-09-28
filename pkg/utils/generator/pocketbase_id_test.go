package generator

import (
	"strings"
	"testing"
)

func TestGeneratePocketBaseID(t *testing.T) {
	// Test basic functionality
	id1 := GeneratePocketBaseID()
	id2 := GeneratePocketBaseID()

	// Verify IDs are generated
	if id1 == "" {
		t.Error("Generated ID should not be empty")
	}

	if id2 == "" {
		t.Error("Generated ID should not be empty")
	}

	// Verify IDs are unique
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}

	// Verify ID format (hex string with prefix)
	if len(id1) != 15 {
		t.Errorf("Expected ID length of 15 characters, got %d", len(id1))
	}

	// Verify ID starts with correct prefix
	if id1[0] != 'r' {
		t.Errorf("Expected ID to start with 'r', got %c", id1[0])
	}

	// Verify all characters after prefix are valid hex characters
	for i := 1; i < len(id1); i++ {
		char := id1[i]
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("ID contains invalid character: %c", char)
		}
	}

	// Test multiple generations for uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GeneratePocketBaseID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true

		if len(id) != 15 {
			t.Errorf("ID %d has incorrect length: %d", i, len(id))
		}

		// Verify ID format
		if id[0] != 'r' {
			t.Errorf("ID %d should start with 'r': %s", i, id)
		}
	}

	// Verify we generated 100 unique IDs
	if len(ids) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(ids))
	}
}

func TestGeneratePocketBaseIDConsistency(t *testing.T) {
	// Test that the function is consistent in its behavior
	// Generate a few IDs and verify they follow the expected pattern

	for i := 0; i < 10; i++ {
		id := GeneratePocketBaseID()

		// Should always be 15 characters
		if len(id) != 15 {
			t.Errorf("ID %d has incorrect length: %d", i, len(id))
		}

		// Should always start with 'r'
		if id[0] != 'r' {
			t.Errorf("ID %d should start with 'r': %s", i, id)
		}

		// Should always be lowercase hex after prefix
		for j := 1; j < len(id); j++ {
			char := id[j]
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("ID %d contains invalid character: %c", i, char)
			}
		}

		// Should not contain uppercase letters
		if strings.ContainsAny(id, "ABCDEF") {
			t.Errorf("ID %d should not contain uppercase letters: %s", i, id)
		}
	}
}

func TestGeneratePocketBaseIDEdgeCases(t *testing.T) {
	// Test edge cases and ensure robustness

	// Generate many IDs to test for collisions (very unlikely but good to test)
	ids := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := GeneratePocketBaseID()
		if ids[id] {
			t.Errorf("Collision detected at iteration %d: %s", i, id)
		}
		ids[id] = true
	}

	// Verify all IDs were unique
	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}

	// Verify no ID is empty
	for id := range ids {
		if id == "" {
			t.Error("Empty ID found in generated set")
		}
	}
}

func BenchmarkGeneratePocketBaseID(b *testing.B) {
	// Benchmark the ID generation performance
	for i := 0; i < b.N; i++ {
		GeneratePocketBaseID()
	}
}

func BenchmarkGeneratePocketBaseIDParallel(b *testing.B) {
	// Benchmark parallel ID generation to test thread safety
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GeneratePocketBaseID()
		}
	})
}
