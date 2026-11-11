package zed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type Adapter struct {
	homeDir      string
	workspaceDir string
}

func New(homeDir, workspaceDir string) *Adapter {
	return &Adapter{homeDir: homeDir, workspaceDir: workspaceDir}
}

func (a *Adapter) ID() domain.EditorID {
	return domain.Zed
}

func (a *Adapter) DisplayName() string {
	return "Zed"
}

func (a *Adapter) Detect(ctx context.Context) (*domain.EditorInfo, error) {
	info := &domain.EditorInfo{
		ID:          domain.Zed,
		DisplayName: "Zed",
	}

	configDir := a.configDir()
	settingsPath := filepath.Join(configDir, "settings.json")

	info.ConfigDir = configDir
	info.SettingsPath = settingsPath

	if _, err := os.Stat(configDir); err == nil {
		info.Installed = true
	}

	return info, nil
}

func (a *Adapter) ReadSettings(ctx context.Context, scope domain.SettingsScope) (*domain.Settings, error) {
	path := a.SettingsPath(scope)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &domain.Settings{
				Editor: domain.Zed,
				Scope:  scope,
				Path:   path,
				Data:   make(map[string]any),
			}, nil
		}
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	cleaned := stripJSONComments(data)

	var settings map[string]any
	if err := json.Unmarshal(cleaned, &settings); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}

	return &domain.Settings{
		Editor: domain.Zed,
		Scope:  scope,
		Path:   path,
		Data:   settings,
	}, nil
}

func (a *Adapter) WriteSettings(ctx context.Context, scope domain.SettingsScope, data []byte) error {
	path := a.SettingsPath(scope)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temp settings: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing settings: %w", err)
	}

	return nil
}

func (a *Adapter) SettingsPath(scope domain.SettingsScope) string {
	if scope == domain.ScopeWorkspace && a.workspaceDir != "" {
		return filepath.Join(a.workspaceDir, ".zed", "settings.json")
	}
	return filepath.Join(a.configDir(), "settings.json")
}

func (a *Adapter) ListExtensions(ctx context.Context) ([]domain.Extension, error) {
	extDir := a.extensionsDir()
	entries, err := os.ReadDir(extDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading extensions directory: %w", err)
	}

	var extensions []domain.Extension
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		extensions = append(extensions, domain.Extension{
			ID:   entry.Name(),
			Name: entry.Name(),
		})
	}
	return extensions, nil
}

func (a *Adapter) configDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(a.homeDir, "Library", "Application Support", "Zed")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(a.homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Zed")
	default:
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			configHome = filepath.Join(a.homeDir, ".config")
		}
		return filepath.Join(configHome, "zed")
	}
}

func (a *Adapter) extensionsDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(a.homeDir, "Library", "Application Support", "Zed", "extensions")
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(a.homeDir, "AppData", "Local")
		}
		return filepath.Join(localAppData, "Zed", "extensions")
	default:
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(a.homeDir, ".local", "share")
		}
		return filepath.Join(dataHome, "zed", "extensions")
	}
}

// stripJSONComments removes // line comments from JSON (Zed supports them).
func stripJSONComments(data []byte) []byte {
	var result []byte
	inString := false
	escaped := false

	for i := 0; i < len(data); i++ {
		b := data[i]

		if escaped {
			result = append(result, b)
			escaped = false
			continue
		}

		if inString {
			result = append(result, b)
			if b == '\\' {
				escaped = true
			} else if b == '"' {
				inString = false
			}
			continue
		}

		if b == '"' {
			inString = true
			result = append(result, b)
			continue
		}

		if b == '/' && i+1 < len(data) && data[i+1] == '/' {
			for i < len(data) && data[i] != '\n' {
				i++
			}
			if i < len(data) {
				result = append(result, '\n')
			}
			continue
		}

		result = append(result, b)
	}

	return result
}
