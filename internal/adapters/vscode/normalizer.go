package vscode

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

	if v, ok := settings.Data["editor.formatOnSave"]; ok {
		if b, ok := v.(bool); ok {
			canonical.Editor.FormatOnSave = b
			items = append(items, domain.MigrationItem{
				SourceKey:   "editor.formatOnSave",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.tabSize"]; ok {
		if f, ok := v.(float64); ok {
			canonical.Editor.TabSize = int(f)
			items = append(items, domain.MigrationItem{
				SourceKey:   "editor.tabSize",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.fontSize"]; ok {
		if f, ok := v.(float64); ok {
			canonical.Editor.FontSize = int(f)
			items = append(items, domain.MigrationItem{
				SourceKey:   "editor.fontSize",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.fontFamily"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.FontFamily = s
			items = append(items, domain.MigrationItem{
				SourceKey:   "editor.fontFamily",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.wordWrap"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.WordWrap = s == "on" || s == "wordWrapColumn" || s == "bounded"
			items = append(items, domain.MigrationItem{
				SourceKey:   "editor.wordWrap",
				SourceValue: v,
				Status:      domain.SupportedTransformed,
				Note:        "VS Code word wrap modes simplified to boolean",
			})
		}
	}

	if v, ok := settings.Data["editor.lineNumbers"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.LineNumbers = s
			items = append(items, domain.MigrationItem{
				SourceKey:   "editor.lineNumbers",
				SourceValue: v,
				Status:      domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.minimap.enabled"]; ok {
		items = append(items, domain.MigrationItem{
			SourceKey:   "editor.minimap.enabled",
			SourceValue: v,
			Status:      domain.Unsupported,
			Note:        "Minimap configuration varies by target editor",
		})
	}

	for key, val := range settings.Data {
		if !isHandledKey(key) {
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
		settings["editor.formatOnSave"] = true
	}
	if canonical.Editor.TabSize > 0 {
		settings["editor.tabSize"] = canonical.Editor.TabSize
	}
	if canonical.Editor.FontSize > 0 {
		settings["editor.fontSize"] = canonical.Editor.FontSize
	}
	if canonical.Editor.FontFamily != "" {
		settings["editor.fontFamily"] = canonical.Editor.FontFamily
	}
	if canonical.Editor.WordWrap {
		settings["editor.wordWrap"] = "on"
	}
	if canonical.Editor.LineNumbers != "" {
		settings["editor.lineNumbers"] = canonical.Editor.LineNumbers
	}

	return settings, items, nil
}

var handledKeys = map[string]bool{
	"editor.formatOnSave":    true,
	"editor.tabSize":         true,
	"editor.fontSize":        true,
	"editor.fontFamily":      true,
	"editor.wordWrap":        true,
	"editor.lineNumbers":     true,
	"editor.minimap.enabled": true,
}

func isHandledKey(key string) bool {
	return handledKeys[key]
}
