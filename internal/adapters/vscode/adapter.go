package vscode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	return domain.VSCode
}

func (a *Adapter) DisplayName() string {
	return "Visual Studio Code"
}

func (a *Adapter) Detect(ctx context.Context) (*domain.EditorInfo, error) {
	info := &domain.EditorInfo{
		ID:          domain.VSCode,
		DisplayName: "Visual Studio Code",
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
				Editor: domain.VSCode,
				Scope:  scope,
				Path:   path,
				Data:   make(map[string]any),
			}, nil
		}
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}

	return &domain.Settings{
		Editor: domain.VSCode,
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
		return filepath.Join(a.workspaceDir, ".vscode", "settings.json")
	}
	return filepath.Join(a.configDir(), "settings.json")
}

func (a *Adapter) ListExtensions(ctx context.Context) ([]domain.Extension, error) {
	cmd := exec.CommandContext(ctx, "code", "--list-extensions", "--show-versions")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing extensions: %w", err)
	}

	var extensions []domain.Extension
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "@", 2)
		ext := domain.Extension{ID: parts[0]}
		if len(parts) == 2 {
			ext.Version = parts[1]
		}
		extensions = append(extensions, ext)
	}
	return extensions, nil
}

func (a *Adapter) InstallExtension(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, "code", "--install-extension", id, "--force")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("installing extension %s: %s: %w", id, string(out), err)
	}
	return nil
}

func (a *Adapter) UninstallExtension(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, "code", "--uninstall-extension", id)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("uninstalling extension %s: %s: %w", id, string(out), err)
	}
	return nil
}

func (a *Adapter) configDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(a.homeDir, "Library", "Application Support", "Code", "User")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(a.homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User")
	default:
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			configHome = filepath.Join(a.homeDir, ".config")
		}
		return filepath.Join(configHome, "Code", "User")
	}
}
