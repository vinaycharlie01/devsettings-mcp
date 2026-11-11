package application

import (
	"context"
	"encoding/json"
	"testing"

	sqlitestore "github.com/vinaycharlie01/devsettings-mcp/internal/adapters/sqlite"
	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type mockEditor struct {
	id         domain.EditorID
	name       string
	settings   map[string]any
	extensions []domain.Extension
}

func (m *mockEditor) ID() domain.EditorID { return m.id }
func (m *mockEditor) DisplayName() string { return m.name }
func (m *mockEditor) Detect(_ context.Context) (*domain.EditorInfo, error) {
	return &domain.EditorInfo{ID: m.id, DisplayName: m.name, Installed: true}, nil
}
func (m *mockEditor) ReadSettings(_ context.Context, _ domain.SettingsScope) (*domain.Settings, error) {
	return &domain.Settings{
		Editor: m.id, Scope: domain.ScopeUser,
		Path: "/mock/" + string(m.id) + "/settings.json",
		Data: m.settings,
	}, nil
}
func (m *mockEditor) WriteSettings(_ context.Context, _ domain.SettingsScope, data []byte) error {
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	m.settings = parsed
	return nil
}
func (m *mockEditor) SettingsPath(_ domain.SettingsScope) string {
	return "/mock/" + string(m.id) + "/settings.json"
}
func (m *mockEditor) ListExtensions(_ context.Context) ([]domain.Extension, error) {
	return m.extensions, nil
}

type mockNormalizer struct {
	canonical   *domain.CanonicalSettings
	fromResult  map[string]any
	sourceItems []domain.MigrationItem
	targetItems []domain.MigrationItem
}

func (m *mockNormalizer) ToCanonical(_ context.Context, _ *domain.Settings) (*domain.CanonicalSettings, []domain.MigrationItem, error) {
	return m.canonical, m.sourceItems, nil
}

func (m *mockNormalizer) FromCanonical(_ context.Context, _ *domain.CanonicalSettings) (map[string]any, []domain.MigrationItem, error) {
	return m.fromResult, m.targetItems, nil
}

func setupMigrationTest(t *testing.T) (*MigrationService, *mockEditor, *mockEditor) {
	t.Helper()

	store, err := sqlitestore.New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	source := &mockEditor{
		id: domain.VSCode, name: "VS Code",
		settings: map[string]any{
			"editor.formatOnSave": true,
			"editor.tabSize":      float64(4),
		},
	}
	target := &mockEditor{
		id: domain.Zed, name: "Zed",
		settings: map[string]any{
			"tab_size": float64(2),
		},
	}

	editors := map[domain.EditorID]ports.Editor{
		domain.VSCode: source,
		domain.Zed:    target,
	}

	sourceNorm := &mockNormalizer{
		canonical: &domain.CanonicalSettings{
			Editor: domain.CanonicalEditor{
				FormatOnSave: true,
				TabSize:      4,
			},
			Languages: make(map[string]domain.CanonicalLanguage),
		},
		sourceItems: []domain.MigrationItem{
			{SourceKey: "editor.formatOnSave", SourceValue: true, Status: domain.FullySupported},
			{SourceKey: "editor.tabSize", SourceValue: float64(4), Status: domain.FullySupported},
		},
	}

	targetNorm := &mockNormalizer{
		fromResult: map[string]any{
			"format_on_save": "on",
			"tab_size":       4,
		},
		targetItems: []domain.MigrationItem{
			{TargetKey: "format_on_save", TargetValue: "on", Status: domain.SupportedTransformed},
		},
	}

	normalizers := map[domain.EditorID]ports.SettingsNormalizer{
		domain.VSCode: sourceNorm,
		domain.Zed:    targetNorm,
	}

	svc := NewMigrationService(editors, normalizers, nil, store)
	return svc, source, target
}

func TestPlanMigration(t *testing.T) {
	svc, _, _ := setupMigrationTest(t)
	ctx := context.Background()

	plan, err := svc.PlanMigration(ctx, domain.VSCode, domain.Zed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Source != domain.VSCode {
		t.Errorf("source: got %s, want %s", plan.Source, domain.VSCode)
	}
	if plan.Target != domain.Zed {
		t.Errorf("target: got %s, want %s", plan.Target, domain.Zed)
	}
	if len(plan.Items) == 0 {
		t.Error("expected migration items, got none")
	}
	if plan.ID == "" {
		t.Error("expected plan ID to be set")
	}
}

func TestPlanMigrationUnsupportedEditor(t *testing.T) {
	svc, _, _ := setupMigrationTest(t)
	ctx := context.Background()

	_, err := svc.PlanMigration(ctx, "sublime", domain.Zed)
	if err == nil {
		t.Error("expected error for unsupported editor")
	}
}

func TestPlanMigrationMissingNormalizer(t *testing.T) {
	store, err := sqlitestore.New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	source := &mockEditor{id: domain.VSCode, name: "VS Code", settings: map[string]any{}}
	target := &mockEditor{id: domain.Cursor, name: "Cursor", settings: map[string]any{}}

	editors := map[domain.EditorID]ports.Editor{
		domain.VSCode: source,
		domain.Cursor: target,
	}
	normalizers := map[domain.EditorID]ports.SettingsNormalizer{
		domain.VSCode: &mockNormalizer{
			canonical: &domain.CanonicalSettings{Languages: make(map[string]domain.CanonicalLanguage)},
		},
	}

	svc := NewMigrationService(editors, normalizers, nil, store)
	ctx := context.Background()

	_, err = svc.PlanMigration(ctx, domain.VSCode, domain.Cursor)
	if err == nil {
		t.Error("expected error for missing target normalizer")
	}
}

func TestExecuteMigration(t *testing.T) {
	svc, _, target := setupMigrationTest(t)
	ctx := context.Background()

	plan, err := svc.PlanMigration(ctx, domain.VSCode, domain.Zed)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	err = svc.ExecuteMigration(ctx, plan.ID)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if target.settings == nil {
		t.Fatal("target settings not written")
	}
	if _, ok := target.settings["format_on_save"]; !ok {
		t.Error("expected format_on_save in target settings")
	}
}

func TestExecuteMigrationUnknownPlan(t *testing.T) {
	svc, _, _ := setupMigrationTest(t)
	ctx := context.Background()

	err := svc.ExecuteMigration(ctx, "nonexistent-plan")
	if err == nil {
		t.Error("expected error for unknown plan ID")
	}
}
