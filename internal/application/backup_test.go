package application

import (
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want map[string]any
	}{
		{
			name: "redact api key by key name",
			data: map[string]any{
				"api_key": "sk-1234567890abcdef",
				"name":    "my-project",
			},
			want: map[string]any{
				"api_key": "[REDACTED]",
				"name":    "my-project",
			},
		},
		{
			name: "redact password by key name",
			data: map[string]any{
				"password": "hunter2",
				"host":     "localhost",
			},
			want: map[string]any{
				"password": "[REDACTED]",
				"host":     "localhost",
			},
		},
		{
			name: "redact token by key name",
			data: map[string]any{
				"auth_token": "ghp_xxxxxxxxxxxxxxxxxxxx",
				"port":       8080,
			},
			want: map[string]any{
				"auth_token": "[REDACTED]",
				"port":       8080,
			},
		},
		{
			name: "redact nested secrets",
			data: map[string]any{
				"database": map[string]any{
					"password": "secret123",
					"host":     "localhost",
				},
			},
			want: map[string]any{
				"database": map[string]any{
					"password": "[REDACTED]",
					"host":     "localhost",
				},
			},
		},
		{
			name: "redact AWS key in value",
			data: map[string]any{
				"config": "AKIAIOSFODNN7EXAMPLE",
			},
			want: map[string]any{
				"config": "[REDACTED]",
			},
		},
		{
			name: "no redaction for safe values",
			data: map[string]any{
				"editor.tabSize":      4,
				"editor.formatOnSave": true,
				"editor.fontFamily":   "Fira Code",
			},
			want: map[string]any{
				"editor.tabSize":      4,
				"editor.formatOnSave": true,
				"editor.fontFamily":   "Fira Code",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSecrets(tt.data)
			assertMapsEqual(t, tt.want, got)
		})
	}
}

func TestIsSecretKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"api_key", true},
		{"apikey", true},
		{"API_KEY", true},
		{"password", true},
		{"secret", true},
		{"auth_token", true},
		{"private_key", true},
		{"credentials", true},
		{"editor.tabSize", false},
		{"formatOnSave", false},
		{"fontFamily", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := isSecretKey(tt.key)
			if got != tt.want {
				t.Errorf("isSecretKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
