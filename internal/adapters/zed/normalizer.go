package zed

import (
	"context"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) ToCanonical(_ context.Context, settings *domain.Settings) (*domain.CanonicalSettings, []domain.MigrationItem, error) {
	canonical := &domain.CanonicalSettings{
		Languages: make(map[string]domain.CanonicalLanguage),
	}
	var items []domain.MigrationItem

	if v, ok := settings.Data["format_on_save"]; ok {
		switch val := v.(type) {
		case string:
			canonical.Editor.FormatOnSave = val == "on"
		case bool:
			canonical.Editor.FormatOnSave = val
		}
		items = append(items, domain.MigrationItem{
			SourceKey:   "format_on_save",
			SourceValue: v,
			Status:      domain.FullySupported,
		})
	}

	if v, ok := settings.Data["tab_size"]; ok {
		if f, ok := v.(float64); ok {
			canonical.Editor.TabSize = int(f)
			items = append(items, domain.MigrationItem{
				SourceKey:   "tab_size",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["buffer_font_size"]; ok {
		if f, ok := v.(float64); ok {
			canonical.Editor.FontSize = int(f)
			items = append(items, domain.MigrationItem{
				SourceKey:   "buffer_font_size",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["buffer_font_family"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.FontFamily = s
			items = append(items, domain.MigrationItem{
				SourceKey:   "buffer_font_family",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["vim_mode"]; ok {
		if b, ok := v.(bool); ok {
			canonical.Editor.VimMode = b
			items = append(items, domain.MigrationItem{
				SourceKey:   "vim_mode",
				SourceValue: v,
				Status:      domain.SupportedTransformed,
				Note:        "Zed built-in vim_mode maps to VS Code vim extension",
			})
		}
	}

	if v, ok := settings.Data["theme"]; ok {
		if s, ok := v.(string); ok {
			canonical.Theme = s
			items = append(items, domain.MigrationItem{
				SourceKey:   "theme",
				SourceValue: v,
				Status:      domain.ManualIntervention,
				Note:        "Theme names differ between editors",
			})
		}
	}

	for key, val := range settings.Data {
		if !isHandledZedKey(key) {
			items = append(items, domain.MigrationItem{
				SourceKey:   key,
				SourceValue: val,
				Status:      domain.Unsupported,
				Note:        "No canonical mapping defined",
			})
		}
	}

	return canonical, items, nil
}

func (n *Normalizer) FromCanonical(_ context.Context, canonical *domain.CanonicalSettings) (map[string]any, []domain.MigrationItem, error) {
	settings := make(map[string]any)
	var items []domain.MigrationItem

	if canonical.Editor.FormatOnSave {
		settings["format_on_save"] = "on"
		items = append(items, domain.MigrationItem{
			TargetKey:   "format_on_save",
			TargetValue: "on",
			Status:      domain.SupportedTransformed,
			Note:        "Boolean converted to Zed string format",
		})
	}

	if canonical.Editor.TabSize > 0 {
		settings["tab_size"] = canonical.Editor.TabSize
	}

	if canonical.Editor.FontSize > 0 {
		settings["buffer_font_size"] = canonical.Editor.FontSize
	}

	if canonical.Editor.FontFamily != "" {
		settings["buffer_font_family"] = canonical.Editor.FontFamily
	}

	if canonical.Editor.VimMode {
		settings["vim_mode"] = true
	}

	if canonical.Editor.WordWrap {
		settings["soft_wrap"] = "editor_width"
		items = append(items, domain.MigrationItem{
			TargetKey:   "soft_wrap",
			TargetValue: "editor_width",
			Status:      domain.SupportedTransformed,
			Note:        "Word wrap mapped to Zed soft_wrap",
		})
	}

	return settings, items, nil
}

var handledZedKeys = map[string]bool{
	"format_on_save":     true,
	"tab_size":           true,
	"buffer_font_size":   true,
	"buffer_font_family": true,
	"vim_mode":           true,
	"theme":              true,
}

func isHandledZedKey(key string) bool {
	return handledZedKeys[key]
}
