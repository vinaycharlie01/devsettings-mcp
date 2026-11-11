package zed

import (
	"context"
	"testing"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func TestNormalizerToCanonical(t *testing.T) {
	n := NewNormalizer()

	settings := &domain.Settings{
		Editor: domain.Zed,
		Scope:  domain.ScopeUser,
		Data: map[string]any{
			"format_on_save":     "on",
			"tab_size":           float64(4),
			"buffer_font_size":   float64(14),
			"buffer_font_family": "JetBrains Mono",
			"vim_mode":           true,
			"theme":              "One Dark",
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
	if !canonical.Editor.VimMode {
		t.Error("VimMode should be true")
	}
	if canonical.Theme != "One Dark" {
		t.Errorf("Theme = %q, want %q", canonical.Theme, "One Dark")
	}

	if len(items) == 0 {
		t.Error("expected migration items")
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
			VimMode:      true,
			WordWrap:     true,
		},
	}

	settings, items, err := n.FromCanonical(context.Background(), canonical)
	if err != nil {
		t.Fatal(err)
	}

	if settings["format_on_save"] != "on" {
		t.Errorf("format_on_save = %v, want on", settings["format_on_save"])
	}
	if settings["tab_size"] != 2 {
		t.Errorf("tab_size = %v, want 2", settings["tab_size"])
	}
	if settings["buffer_font_size"] != 16 {
		t.Errorf("buffer_font_size = %v, want 16", settings["buffer_font_size"])
	}
	if settings["buffer_font_family"] != "Fira Code" {
		t.Errorf("buffer_font_family = %v, want Fira Code", settings["buffer_font_family"])
	}
	if settings["vim_mode"] != true {
		t.Errorf("vim_mode = %v, want true", settings["vim_mode"])
	}
	if settings["soft_wrap"] != "editor_width" {
		t.Errorf("soft_wrap = %v, want editor_width", settings["soft_wrap"])
	}

	transformed := 0
	for _, item := range items {
		if item.Status == domain.SupportedTransformed {
			transformed++
		}
	}
	if transformed < 2 {
		t.Errorf("expected at least 2 transformed items, got %d", transformed)
	}
}

func TestFormatOnSaveBooleanInput(t *testing.T) {
	n := NewNormalizer()

	settings := &domain.Settings{
		Editor: domain.Zed,
		Data: map[string]any{
			"format_on_save": true,
		},
	}

	canonical, _, err := n.ToCanonical(context.Background(), settings)
	if err != nil {
		t.Fatal(err)
	}

	if !canonical.Editor.FormatOnSave {
		t.Error("FormatOnSave should handle boolean true")
	}
}
