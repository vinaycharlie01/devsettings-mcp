package cursor

import (
	"context"
	"testing"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func TestNormalizerToCanonical(t *testing.T) {
	n := NewNormalizer()
	ctx := context.Background()

	tests := []struct {
		name         string
		data         map[string]any
		wantFormat   bool
		wantTab      int
		wantFontSize int
		wantMinItems int
	}{
		{
			name: "basic editor settings",
			data: map[string]any{
				"editor.formatOnSave": true,
				"editor.tabSize":      float64(4),
				"editor.fontSize":     float64(14),
			},
			wantFormat:   true,
			wantTab:      4,
			wantFontSize: 14,
			wantMinItems: 3,
		},
		{
			name: "cursor AI settings tracked as unsupported",
			data: map[string]any{
				"editor.tabSize":     float64(2),
				"cursor.tab.enabled": true,
				"cursor.ai.model":    "gpt-4",
			},
			wantTab:      2,
			wantMinItems: 3,
		},
		{
			name:         "empty settings",
			data:         map[string]any{},
			wantMinItems: 0,
		},
		{
			name: "word wrap on",
			data: map[string]any{
				"editor.wordWrap": "on",
			},
			wantMinItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := &domain.Settings{Editor: domain.Cursor, Data: tt.data}
			canonical, items, err := n.ToCanonical(ctx, settings)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if canonical.Editor.FormatOnSave != tt.wantFormat {
				t.Errorf("FormatOnSave: got %v, want %v", canonical.Editor.FormatOnSave, tt.wantFormat)
			}
			if canonical.Editor.TabSize != tt.wantTab {
				t.Errorf("TabSize: got %d, want %d", canonical.Editor.TabSize, tt.wantTab)
			}
			if canonical.Editor.FontSize != tt.wantFontSize {
				t.Errorf("FontSize: got %d, want %d", canonical.Editor.FontSize, tt.wantFontSize)
			}
			if len(items) < tt.wantMinItems {
				t.Errorf("items: got %d, want at least %d", len(items), tt.wantMinItems)
			}
		})
	}
}

func TestNormalizerToCanonicalCursorSpecific(t *testing.T) {
	n := NewNormalizer()
	ctx := context.Background()

	settings := &domain.Settings{
		Editor: domain.Cursor,
		Data: map[string]any{
			"cursor.tab.enabled":        true,
			"cursor.chat.enabled":       true,
			"cursor.composer.enabled":   false,
			"cursor.prediction.enabled": true,
		},
	}

	_, items, err := n.ToCanonical(ctx, settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	unsupported := 0
	for _, item := range items {
		if item.Status == domain.Unsupported {
			unsupported++
		}
	}
	if unsupported != 4 {
		t.Errorf("expected 4 unsupported Cursor-specific items, got %d", unsupported)
	}
}

func TestNormalizerFromCanonical(t *testing.T) {
	n := NewNormalizer()
	ctx := context.Background()

	tests := []struct {
		name      string
		canonical *domain.CanonicalSettings
		wantKeys  []string
	}{
		{
			name: "basic settings",
			canonical: &domain.CanonicalSettings{
				Editor: domain.CanonicalEditor{
					FormatOnSave: true,
					TabSize:      4,
					FontSize:     14,
				},
			},
			wantKeys: []string{"editor.formatOnSave", "editor.tabSize", "editor.fontSize"},
		},
		{
			name: "with word wrap and font family",
			canonical: &domain.CanonicalSettings{
				Editor: domain.CanonicalEditor{
					WordWrap:   true,
					FontFamily: "Fira Code",
				},
			},
			wantKeys: []string{"editor.wordWrap", "editor.fontFamily"},
		},
		{
			name: "vim mode triggers manual intervention",
			canonical: &domain.CanonicalSettings{
				Editor: domain.CanonicalEditor{
					VimMode: true,
				},
			},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, _, err := n.FromCanonical(ctx, tt.canonical)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, key := range tt.wantKeys {
				if _, ok := settings[key]; !ok {
					t.Errorf("missing expected key %q", key)
				}
			}
		})
	}
}
