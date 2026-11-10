package vscode

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

	if info.ID != domain.VSCode {
		t.Errorf("ID = %q, want %q", info.ID, domain.VSCode)
	}
	if info.Installed {
		t.Error("should not be installed in temp dir")
	}

	configDir := info.ConfigDir
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	info, err = adapter.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !info.Installed {
		t.Error("should be detected as installed after creating config dir")
	}
}

func TestReadSettings(t *testing.T) {
	homeDir := t.TempDir()
	adapter := New(homeDir, "")

	t.Run("nonexistent returns empty", func(t *testing.T) {
		settings, err := adapter.ReadSettings(context.Background(), domain.ScopeUser)
		if err != nil {
			t.Fatal(err)
		}
		if len(settings.Data) != 0 {
			t.Errorf("expected empty settings, got %d keys", len(settings.Data))
		}
	})

	t.Run("reads existing settings", func(t *testing.T) {
		settingsPath := adapter.SettingsPath(domain.ScopeUser)
		if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
			t.Fatal(err)
		}

		data := map[string]any{
			"editor.tabSize":      float64(4),
			"editor.formatOnSave": true,
		}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		if err := os.WriteFile(settingsPath, jsonData, 0o644); err != nil {
			t.Fatal(err)
		}

		settings, err := adapter.ReadSettings(context.Background(), domain.ScopeUser)
		if err != nil {
			t.Fatal(err)
		}

		if settings.Data["editor.formatOnSave"] != true {
			t.Error("formatOnSave should be true")
		}
		if settings.Data["editor.tabSize"] != float64(4) {
			t.Errorf("tabSize = %v, want 4", settings.Data["editor.tabSize"])
		}
	})
}

func TestWriteSettings(t *testing.T) {
	homeDir := t.TempDir()
	adapter := New(homeDir, "")

	data := map[string]any{"editor.tabSize": 4}
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

	if got["editor.tabSize"] != float64(4) {
		t.Errorf("tabSize = %v, want 4", got["editor.tabSize"])
	}
}

func TestSettingsPath(t *testing.T) {
	homeDir := "/home/testuser"
	workspace := "/projects/myapp"
	adapter := New(homeDir, workspace)

	userPath := adapter.SettingsPath(domain.ScopeUser)
	if !filepath.IsAbs(userPath) {
		t.Errorf("user path should be absolute: %s", userPath)
	}

	workspacePath := adapter.SettingsPath(domain.ScopeWorkspace)
	expected := filepath.Join(workspace, ".vscode", "settings.json")
	if workspacePath != expected {
		t.Errorf("workspace path = %q, want %q", workspacePath, expected)
	}
}

func TestConfigDir(t *testing.T) {
	adapter := New("/home/testuser", "")
	configDir := adapter.configDir()

	switch runtime.GOOS {
	case "linux":
		expected := "/home/testuser/.config/Code/User"
		if configDir != expected {
			t.Errorf("configDir = %q, want %q", configDir, expected)
		}
	case "darwin":
		expected := "/home/testuser/Library/Application Support/Code/User"
		if configDir != expected {
			t.Errorf("configDir = %q, want %q", configDir, expected)
		}
	}
}
