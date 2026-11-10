package cursor

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
				SourceKey: "editor.formatOnSave", SourceValue: v, Status: domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.tabSize"]; ok {
		if f, ok := v.(float64); ok {
			canonical.Editor.TabSize = int(f)
			items = append(items, domain.MigrationItem{
				SourceKey: "editor.tabSize", SourceValue: v, Status: domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.fontSize"]; ok {
		if f, ok := v.(float64); ok {
			canonical.Editor.FontSize = int(f)
			items = append(items, domain.MigrationItem{
				SourceKey: "editor.fontSize", SourceValue: v, Status: domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.fontFamily"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.FontFamily = s
			items = append(items, domain.MigrationItem{
				SourceKey: "editor.fontFamily", SourceValue: v, Status: domain.FullySupported,
			})
		}
	}

	if v, ok := settings.Data["editor.wordWrap"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.WordWrap = s == "on" || s == "wordWrapColumn" || s == "bounded"
			items = append(items, domain.MigrationItem{
				SourceKey: "editor.wordWrap", SourceValue: v, Status: domain.SupportedTransformed,
				Note: "VS Code word wrap modes simplified to boolean",
			})
		}
	}

	if v, ok := settings.Data["editor.lineNumbers"]; ok {
		if s, ok := v.(string); ok {
			canonical.Editor.LineNumbers = s
			items = append(items, domain.MigrationItem{
				SourceKey: "editor.lineNumbers", SourceValue: v, Status: domain.FullySupported,
			})
		}
	}

	// AI settings: map to canonical model
	if v, ok := settings.Data["cursor.tab.enabled"]; ok {
		if b, ok := v.(bool); ok {
			canonical.AI.InlineCompletion = b
			items = append(items, domain.MigrationItem{
				SourceKey: "cursor.tab.enabled", SourceValue: v,
				Status: domain.SupportedTransformed,
				Note:   "Maps to inline completion in target editor",
			})
		}
	}

	if v, ok := settings.Data["cursor.chat.enabled"]; ok {
		if b, ok := v.(bool); ok {
			canonical.AI.ChatEnabled = b
			items = append(items, domain.MigrationItem{
				SourceKey: "cursor.chat.enabled", SourceValue: v,
				Status: domain.SupportedTransformed,
				Note:   "Maps to AI chat in target editor",
			})
		}
	}

	if v, ok := settings.Data["cursor.ai.model"]; ok {
		if s, ok := v.(string); ok {
			canonical.AI.Model = s
			canonical.AI.Provider = "cursor"
			items = append(items, domain.MigrationItem{
				SourceKey: "cursor.ai.model", SourceValue: v,
				Status: domain.PartiallySupported,
				Note:   "Model name may differ in target editor",
			})
		}
	}

	if v, ok := settings.Data["cursor.composer.enabled"]; ok {
		if b, ok := v.(bool); ok {
			canonical.AI.ComposerEnabled = b
			items = append(items, domain.MigrationItem{
				SourceKey: "cursor.composer.enabled", SourceValue: v,
				Status: domain.PartiallySupported,
				Note:   "Composer is Cursor-specific; no direct equivalent in other editors",
			})
		}
	}

	if v, ok := settings.Data["cursor.prediction.enabled"]; ok {
		if b, ok := v.(bool); ok {
			canonical.AI.PredictionEnabled = b
			items = append(items, domain.MigrationItem{
				SourceKey: "cursor.prediction.enabled", SourceValue: v,
				Status: domain.PartiallySupported,
				Note:   "Prediction is Cursor-specific; no direct equivalent in other editors",
			})
		}
	}

	if v, ok := settings.Data["cursor.copilot.enabled"]; ok {
		items = append(items, domain.MigrationItem{
			SourceKey: "cursor.copilot.enabled", SourceValue: v,
			Status: domain.PartiallySupported,
			Note:   "Cursor Copilot integration maps to inline completion",
		})
	}

	for key, val := range settings.Data {
		if !isHandledCursorKey(key) {
			items = append(items, domain.MigrationItem{
				SourceKey: key, SourceValue: val,
				Status: domain.Unsupported, Note: "No canonical mapping defined",
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
	if canonical.Editor.VimMode {
		items = append(items, domain.MigrationItem{
			TargetKey: "vim_mode", TargetValue: true,
			Status: domain.ManualIntervention,
			Note:   "Install vscodevim.vim extension for vim support in Cursor",
		})
	}

	// AI settings → Cursor
	if canonical.AI.InlineCompletion {
		settings["cursor.tab.enabled"] = true
		items = append(items, domain.MigrationItem{
			TargetKey: "cursor.tab.enabled", TargetValue: true,
			Status: domain.SupportedTransformed,
			Note:   "Inline completion mapped to Cursor Tab",
		})
	}
	if canonical.AI.ChatEnabled {
		settings["cursor.chat.enabled"] = true
		items = append(items, domain.MigrationItem{
			TargetKey: "cursor.chat.enabled", TargetValue: true,
			Status: domain.SupportedTransformed,
			Note:   "AI chat mapped to Cursor Chat",
		})
	}
	if canonical.AI.Model != "" {
		settings["cursor.ai.model"] = canonical.AI.Model
		items = append(items, domain.MigrationItem{
			TargetKey: "cursor.ai.model", TargetValue: canonical.AI.Model,
			Status: domain.PartiallySupported,
			Note:   "Model name may need manual adjustment for Cursor",
		})
	}
	if canonical.AI.ComposerEnabled {
		settings["cursor.composer.enabled"] = true
	}
	if canonical.AI.PredictionEnabled {
		settings["cursor.prediction.enabled"] = true
	}

	return settings, items, nil
}

var handledCursorKeys = map[string]bool{
	"editor.formatOnSave":       true,
	"editor.tabSize":            true,
	"editor.fontSize":           true,
	"editor.fontFamily":         true,
	"editor.wordWrap":           true,
	"editor.lineNumbers":        true,
	"cursor.tab.enabled":        true,
	"cursor.chat.enabled":       true,
	"cursor.copilot.enabled":    true,
	"cursor.ai.model":           true,
	"cursor.composer.enabled":   true,
	"cursor.prediction.enabled": true,
}

func isHandledCursorKey(key string) bool {
	return handledCursorKeys[key]
}
