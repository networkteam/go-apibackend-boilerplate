package command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"myvendor.mytld/myproject/backend/domain/command"
)

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Trimming whitespace
		{name: "trim leading spaces", input: "  username", expected: "username"},
		{name: "trim trailing spaces", input: "username  ", expected: "username"},
		{name: "trim both sides", input: "  username  ", expected: "username"},
		{name: "trim tabs", input: "\tusername\t", expected: "username"},
		{name: "trim newlines", input: "\nusername\n", expected: "username"},
		{name: "trim mixed whitespace", input: " \t username \n ", expected: "username"},

		// Lowercase conversion
		{name: "uppercase to lowercase", input: "USERNAME", expected: "username"},
		{name: "mixed case to lowercase", input: "UserName", expected: "username"},
		{name: "already lowercase", input: "username", expected: "username"},
		{name: "uppercase with numbers", input: "USER123", expected: "user123"},

		// Combined trimming and lowercasing
		{name: "trim and lowercase", input: "  UserName  ", expected: "username"},
		{name: "complex whitespace and case", input: " \t MixedCase \n ", expected: "mixedcase"},

		// Special characters (preserved, only trim and lowercase)
		{name: "dots preserved", input: "test.user", expected: "test.user"},
		{name: "hyphens preserved", input: "test-user", expected: "test-user"},
		{name: "underscores preserved", input: "test_user", expected: "test_user"},
		{name: "numbers preserved", input: "user123", expected: "user123"},

		// Umlauts (preserved, lowercased)
		{name: "uppercase umlaut ü", input: "MÜLLER", expected: "müller"},
		{name: "uppercase umlaut ö", input: "KÖNIG", expected: "könig"},
		{name: "uppercase umlaut ä", input: "KÄSE", expected: "käse"},
		{name: "mixed case umlauts", input: "Müller", expected: "müller"},
		{name: "lowercase ß preserved", input: "straße", expected: "straße"},

		// Empty and whitespace-only
		{name: "empty string", input: "", expected: ""},
		{name: "only spaces", input: "   ", expected: ""},
		{name: "only tabs", input: "\t\t", expected: ""},

		// Real-world examples
		{name: "typical username", input: " John.Doe ", expected: "john.doe"},
		{name: "username with numbers", input: "User_123", expected: "user_123"},
		{name: "mixed format", input: "  Test-User_01  ", expected: "test-user_01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := command.SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, result, "SanitizeUsername(%q) should return %q, got %q", tt.input, tt.expected, result)
		})
	}
}

func TestSanitizeEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Trimming whitespace
		{name: "trim leading spaces", input: "  email@test.com", expected: "email@test.com"},
		{name: "trim trailing spaces", input: "email@test.com  ", expected: "email@test.com"},
		{name: "trim both sides", input: "  email@test.com  ", expected: "email@test.com"},
		{name: "trim tabs", input: "\temail@test.com\t", expected: "email@test.com"},
		{name: "trim mixed whitespace", input: " \t email@test.com \n ", expected: "email@test.com"},

		// Lowercase conversion
		{name: "uppercase to lowercase", input: "EMAIL@TEST.COM", expected: "email@test.com"},
		{name: "mixed case to lowercase", input: "User@Example.Com", expected: "user@example.com"},
		{name: "already lowercase", input: "test@example.com", expected: "test@example.com"},
		{name: "uppercase domain", input: "test@EXAMPLE.COM", expected: "test@example.com"},

		// Combined trimming and lowercasing
		{name: "trim and lowercase", input: "  TEST@Example.COM  ", expected: "test@example.com"},
		{name: "complex whitespace and case", input: " \t User@Domain.Co.UK \n ", expected: "user@domain.co.uk"},

		// Special email formats (preserved)
		{name: "email with plus", input: "test+tag@example.com", expected: "test+tag@example.com"},
		{name: "email with dots", input: "first.last@example.com", expected: "first.last@example.com"},
		{name: "email with numbers", input: "user123@test456.com", expected: "user123@test456.com"},
		{name: "email with hyphen", input: "test@my-domain.com", expected: "test@my-domain.com"},

		// Empty and whitespace-only
		{name: "empty string", input: "", expected: ""},
		{name: "only spaces", input: "   ", expected: ""},

		// Real-world examples
		{name: "typical email", input: " John.Doe@Example.COM ", expected: "john.doe@example.com"},
		{name: "email with subdomain", input: "USER@MAIL.COMPANY.COM", expected: "user@mail.company.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := command.SanitizeEmailAddress(tt.input)
			assert.Equal(t, tt.expected, result, "SanitizeEmailAddress(%q) should return %q, got %q", tt.input, tt.expected, result)
		})
	}
}

func TestSanitizePassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Trimming whitespace (ONLY operation for passwords)
		{name: "trim leading spaces", input: "  password", expected: "password"},
		{name: "trim trailing spaces", input: "password  ", expected: "password"},
		{name: "trim both sides", input: "  password  ", expected: "password"},
		{name: "trim tabs", input: "\tpassword\t", expected: "password"},
		{name: "trim newlines", input: "\npassword\n", expected: "password"},
		{name: "trim mixed whitespace", input: " \t password \n ", expected: "password"},

		// Case preservation (NOT converted to lowercase)
		{name: "uppercase preserved", input: "PASSWORD", expected: "PASSWORD"},
		{name: "mixed case preserved", input: "Password123", expected: "Password123"},
		{name: "lowercase unchanged", input: "password", expected: "password"},

		// Special characters preserved
		{name: "with special chars", input: "Pass123!", expected: "Pass123!"},
		{name: "complex password", input: "MyP@ssw0rd!2024", expected: "MyP@ssw0rd!2024"},
		{name: "symbols preserved", input: "!@#$%^&*()", expected: "!@#$%^&*()"},

		// Combined trimming with case/special chars preserved
		{name: "trim with uppercase", input: "  PASSWORD123!  ", expected: "PASSWORD123!"},
		{name: "trim with mixed case", input: " \t ValidPass123! \n ", expected: "ValidPass123!"},

		// Empty and whitespace-only
		{name: "empty string", input: "", expected: ""},
		{name: "only spaces", input: "   ", expected: ""},
		{name: "only tabs", input: "\t\t", expected: ""},

		// Real-world examples
		{name: "typical password", input: " MySecret123! ", expected: "MySecret123!"},
		{name: "no changes needed", input: "NoSpaces123!", expected: "NoSpaces123!"},
		{name: "whitespace only trimmed", input: "  P@ssw0rd  ", expected: "P@ssw0rd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := command.SanitizePassword(tt.input)
			assert.Equal(t, tt.expected, result, "SanitizePassword(%q) should return %q, got %q", tt.input, tt.expected, result)
		})
	}
}
