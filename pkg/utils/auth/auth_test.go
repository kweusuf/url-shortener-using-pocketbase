package auth

import (
	"errors"
	"testing"

	"github.com/kweusuf/pocketbase-demo/pkg/constants"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected error
	}{
		{"", errors.New(constants.ErrorEmailRequired)},
		{"invalid", errors.New(constants.ErrorInvalidEmail)},
		{"valid@example.com", nil},
		{"another@domain.org", nil},
		{"@", nil},
		{"user@", nil},
		{"@domain.com", nil},
	}

	for _, test := range tests {
		err := ValidateEmail(test.email)
		if (err == nil && test.expected != nil) || (err != nil && test.expected == nil) || (err != nil && test.expected != nil && err.Error() != test.expected.Error()) {
			t.Errorf("ValidateEmail(%q) = %v; want %v", test.email, err, test.expected)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		expected error
	}{
		{"", errors.New(constants.ErrorPasswordRequired)},
		{"12345", errors.New(constants.ErrorPasswordTooShort)},
		{"123456", nil},
		{"password", nil},
		{"strongpassword", nil},
	}

	for _, test := range tests {
		err := ValidatePassword(test.password)
		if (err == nil && test.expected != nil) || (err != nil && test.expected == nil) || (err != nil && test.expected != nil && err.Error() != test.expected.Error()) {
			t.Errorf("ValidatePassword(%q) = %v; want %v", test.password, err, test.expected)
		}
	}
}

func TestHashPassword(t *testing.T) {
	password := "mypassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword returned empty hash")
	}
	if hash == password {
		t.Error("Hash should not be equal to original password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "mypassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Correct password
	err = VerifyPassword(password, hash)
	if err != nil {
		t.Errorf("VerifyPassword with correct password failed: %v", err)
	}

	// Incorrect password
	err = VerifyPassword("wrongpassword", hash)
	if err == nil {
		t.Error("VerifyPassword with incorrect password should have failed")
	}
}
