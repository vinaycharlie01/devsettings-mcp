package zed

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func TestDetect(t *testing.T) {
	homeDir := t.TempDir()
	adapter := New(homeDir, "")

	info, err := adapter.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if info.ID != domain.Zed {
		t.Errorf("ID = %q, want %q", info.ID, domain.Zed)
	}
	if info.Installed {
		t.Error("should not be installed in temp dir")
	}
}

func TestReadSettingsWithComments(t *testing.T) {
	homeDir := t.TempDir()
	adapter := New(homeDir, "")

	settingsPath := adapter.SettingsPath(domain.ScopeUser)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}

	content := `{
  // This is a comment
  "tab_size": 4,
  "format_on_save": "on" // inline comment
}`
	if err := os.WriteFile(settingsPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, err := adapter.ReadSettings(context.Background(), domain.ScopeUser)
	if err != nil {
		t.Fatal(err)
	}

	if settings.Data["tab_size"] != float64(4) {
		t.Errorf("tab_size = %v, want 4", settings.Data["tab_size"])
	}
	if settings.Data["format_on_save"] != "on" {
		t.Errorf("format_on_save = %v, want on", settings.Data["format_on_save"])
	}
}

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no comments",
			input: `{"key": "value"}`,
			want:  `{"key": "value"}`,
		},
		{
			name:  "line comment",
			input: "// comment\n{\"key\": \"value\"}",
			want:  "\n{\"key\": \"value\"}",
		},
		{
			name:  "inline comment",
			input: `{"key": "value"} // comment`,
			want:  `{"key": "value"} `,
		},
		{
			name:  "comment-like string preserved",
			input: `{"url": "https://example.com"}`,
			want:  `{"url": "https://example.com"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripJSONComments([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListExtensions(t *testing.T) {
	homeDir := t.TempDir()
	adapter := New(homeDir, "")

	extDir := adapter.extensionsDir()
	if err := os.MkdirAll(filepath.Join(extDir, "rust-analyzer"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(extDir, "toml"), 0o755); err != nil {
		t.Fatal(err)
	}

	extensions, err := adapter.ListExtensions(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(extensions) != 2 {
		t.Errorf("got %d extensions, want 2", len(extensions))
	}
}

func TestWriteSettings(t *testing.T) {
	homeDir := t.TempDir()
	adapter := New(homeDir, "")

	data := map[string]any{"tab_size": 4}
	jsonData, _ := json.MarshalIndent(data, "", "  ")

	err := adapter.WriteSettings(context.Background(), domain.ScopeUser, jsonData)
	if err != nil {
		t.Fatal(err)
	}

	readBack, err := os.ReadFile(adapter.SettingsPath(domain.ScopeUser))
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(readBack, &got); err != nil {
		t.Fatal(err)
	}

	if got["tab_size"] != float64(4) {
		t.Errorf("tab_size = %v, want 4", got["tab_size"])
	}
}

func TestConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	adapter := New("/home/testuser", "")
	configDir := adapter.configDir()

	switch runtime.GOOS {
	case "linux":
		expected := "/home/testuser/.config/zed"
		if configDir != expected {
			t.Errorf("configDir = %q, want %q", configDir, expected)
		}
	case "darwin":
		expected := "/home/testuser/Library/Application Support/Zed"
		if configDir != expected {
			t.Errorf("configDir = %q, want %q", configDir, expected)
		}
	}
}
