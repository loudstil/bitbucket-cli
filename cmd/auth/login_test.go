package auth

import (
	"testing"
)

func TestStripControlChars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"hello\x00world", "helloworld"},
		{"token\r\n", "token"},
		{"\x1b[31mred\x1b[0m", "[31mred[0m"}, // ESC (0x1b) stripped, printable chars kept
		{"\x7ftoken", "token"},
		{"clean-token-123", "clean-token-123"},
	}

	for _, tt := range tests {
		got := stripControlChars(tt.input)
		if got != tt.want {
			t.Errorf("stripControlChars(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHostFromURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://bitbucket.example.com", "bitbucket.example.com"},
		{"http://bitbucket.example.com", "bitbucket.example.com"},
		{"https://bitbucket.example.com/", "bitbucket.example.com"},
		{"https://bitbucket.example.com/some/path", "bitbucket.example.com"},
		{"", "datacenter"},
	}

	for _, tt := range tests {
		got := hostFromURL(tt.input)
		if got != tt.want {
			t.Errorf("hostFromURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
