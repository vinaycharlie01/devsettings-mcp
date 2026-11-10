package vscode

import (
	"context"
	"testing"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func TestNormalizerToCanonical(t *testing.T) {
	n := NewNormalizer()

	settings := &domain.Settings{
		Editor: domain.VSCode,
		Scope:  domain.ScopeUser,
		Data: map[string]any{
			"editor.formatOnSave": true,
			"editor.tabSize":      float64(4),
			"editor.fontSize":     float64(14),
			"editor.fontFamily":   "JetBrains Mono",
			"editor.wordWrap":     "on",
			"editor.lineNumbers":  "relative",
		},
	}

	canonical, items, err := n.ToCanonical(context.Background(), settings)
	if err != nil {
		t.Fatal(err)
	}

	if !canonical.Editor.FormatOnSave {
		t.Error("FormatOnSave should be true")
	}
	if canonical.Editor.TabSize != 4 {
		t.Errorf("TabSize = %d, want 4", canonical.Editor.TabSize)
	}
	if canonical.Editor.FontSize != 14 {
		t.Errorf("FontSize = %d, want 14", canonical.Editor.FontSize)
	}
	if canonical.Editor.FontFamily != "JetBrains Mono" {
		t.Errorf("FontFamily = %q, want %q", canonical.Editor.FontFamily, "JetBrains Mono")
	}
	if !canonical.Editor.WordWrap {
		t.Error("WordWrap should be true")
	}
	if canonical.Editor.LineNumbers != "relative" {
		t.Errorf("LineNumbers = %q, want %q", canonical.Editor.LineNumbers, "relative")
	}

	supported := 0
	for _, item := range items {
		if item.Status == domain.FullySupported || item.Status == domain.SupportedTransformed {
			supported++
		}
	}
	if supported != 6 {
		t.Errorf("expected 6 supported items, got %d", supported)
	}
}

func TestNormalizerFromCanonical(t *testing.T) {
	n := NewNormalizer()

	canonical := &domain.CanonicalSettings{
		Editor: domain.CanonicalEditor{
			FormatOnSave: true,
			TabSize:      2,
			FontSize:     16,
			FontFamily:   "Fira Code",
			WordWrap:     true,
			LineNumbers:  "on",
		},
	}

	settings, _, err := n.FromCanonical(context.Background(), canonical)
	if err != nil {
		t.Fatal(err)
	}

	if settings["editor.formatOnSave"] != true {
		t.Error("formatOnSave should be true")
	}
	if settings["editor.tabSize"] != 2 {
		t.Errorf("tabSize = %v, want 2", settings["editor.tabSize"])
	}
	if settings["editor.fontSize"] != 16 {
		t.Errorf("fontSize = %v, want 16", settings["editor.fontSize"])
	}
	if settings["editor.fontFamily"] != "Fira Code" {
		t.Errorf("fontFamily = %v, want Fira Code", settings["editor.fontFamily"])
	}
	if settings["editor.wordWrap"] != "on" {
		t.Errorf("wordWrap = %v, want on", settings["editor.wordWrap"])
	}
}

func TestNormalizerUnsupportedKeys(t *testing.T) {
	n := NewNormalizer()

	settings := &domain.Settings{
		Editor: domain.VSCode,
		Data: map[string]any{
			"editor.formatOnSave":   true,
			"some.custom.setting":   "value",
			"another.unknown":       42,
		},
	}

	_, items, err := n.ToCanonical(context.Background(), settings)
	if err != nil {
		t.Fatal(err)
	}

	unsupported := 0
	for _, item := range items {
		if item.Status == domain.Unsupported {
			unsupported++
		}
	}
	if unsupported != 2 {
		t.Errorf("expected 2 unsupported items, got %d", unsupported)
	}
}
