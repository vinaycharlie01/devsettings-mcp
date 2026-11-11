package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
	"gopkg.in/yaml.v3"
)

type ProfileService struct {
	profileDir string
	editors    map[domain.EditorID]ports.Editor
}

func NewProfileService(profileDir string, editors map[domain.EditorID]ports.Editor) *ProfileService {
	return &ProfileService{profileDir: profileDir, editors: editors}
}

func (s *ProfileService) ListProfiles(_ context.Context) ([]domain.Profile, error) {
	entries, err := os.ReadDir(s.profileDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading profiles directory: %w", err)
	}

	var profiles []domain.Profile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		profile, err := loadProfileFile(filepath.Join(s.profileDir, name))
		if err != nil {
			continue
		}
		profiles = append(profiles, *profile)
	}
	return profiles, nil
}

func (s *ProfileService) LoadProfile(_ context.Context, name string) (*domain.Profile, error) {
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(s.profileDir, name+ext)
		profile, err := loadProfileFile(path)
		if err == nil {
			return profile, nil
		}
	}
	return nil, fmt.Errorf("profile not found: %s", name)
}

func (s *ProfileService) ApplyProfile(ctx context.Context, name string, editorID domain.EditorID) (map[string]any, error) {
	profile, err := s.LoadProfile(ctx, name)
	if err != nil {
		return nil, err
	}

	editor, ok := s.editors[editorID]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", editorID)
	}

	settings := profileToEditorSettings(profile, editorID)

	current, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading current settings: %w", err)
	}

	merged := DeepMerge(current.Data, settings)

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling settings: %w", err)
	}

	if err := editor.WriteSettings(ctx, domain.ScopeUser, data); err != nil {
		return nil, fmt.Errorf("writing settings: %w", err)
	}

	return settings, nil
}

func loadProfileFile(path string) (*domain.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var profile domain.Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile %s: %w", path, err)
	}
	return &profile, nil
}

func profileToEditorSettings(profile *domain.Profile, editorID domain.EditorID) map[string]any {
	settings := make(map[string]any)

	switch editorID {
	case domain.VSCode, domain.Cursor:
		if v, ok := profile.Settings["format_on_save"]; ok {
			settings["editor.formatOnSave"] = v
		}
		if v, ok := profile.Settings["tab_size"]; ok {
			settings["editor.tabSize"] = v
		}
		for lang, cfg := range profile.Languages {
			langKey := fmt.Sprintf("[%s]", lang)
			langSettings := make(map[string]any)
			if cfg.Formatter != "" {
				langSettings["editor.defaultFormatter"] = cfg.Formatter
			}
			if len(langSettings) > 0 {
				settings[langKey] = langSettings
			}
		}
	case domain.Zed:
		if v, ok := profile.Settings["format_on_save"]; ok {
			if b, ok := v.(bool); ok && b {
				settings["format_on_save"] = "on"
			}
		}
		if v, ok := profile.Settings["tab_size"]; ok {
			settings["tab_size"] = v
		}
		langs := make(map[string]any)
		for lang, cfg := range profile.Languages {
			langSettings := make(map[string]any)
			if cfg.Formatter != "" {
				langSettings["formatter"] = cfg.Formatter
			}
			if cfg.LSP != "" {
				langSettings["language_server"] = cfg.LSP
			}
			if len(langSettings) > 0 {
				langs[lang] = langSettings
			}
		}
		if len(langs) > 0 {
			settings["languages"] = langs
		}
	}

	return settings
}
