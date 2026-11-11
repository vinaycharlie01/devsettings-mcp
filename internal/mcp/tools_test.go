package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	sqlitestore "github.com/vinaycharlie01/devsettings-mcp/internal/adapters/sqlite"
	"github.com/vinaycharlie01/devsettings-mcp/internal/application"
	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type testEditor struct {
	id         domain.EditorID
	name       string
	settings   map[string]any
	extensions []domain.Extension
}

func (e *testEditor) ID() domain.EditorID { return e.id }
func (e *testEditor) DisplayName() string { return e.name }
func (e *testEditor) Detect(_ context.Context) (*domain.EditorInfo, error) {
	return &domain.EditorInfo{ID: e.id, DisplayName: e.name, Installed: true}, nil
}
func (e *testEditor) ReadSettings(_ context.Context, _ domain.SettingsScope) (*domain.Settings, error) {
	return &domain.Settings{
		Editor: e.id, Scope: domain.ScopeUser,
		Path: "/test/" + string(e.id) + "/settings.json",
		Data: e.settings,
	}, nil
}
func (e *testEditor) WriteSettings(_ context.Context, _ domain.SettingsScope, data []byte) error {
	var parsed map[string]any
	json.Unmarshal(data, &parsed)
	e.settings = parsed
	return nil
}
func (e *testEditor) SettingsPath(_ domain.SettingsScope) string {
	return "/test/" + string(e.id) + "/settings.json"
}
func (e *testEditor) ListExtensions(_ context.Context) ([]domain.Extension, error) {
	return e.extensions, nil
}

type testNormalizer struct{}

func (n *testNormalizer) ToCanonical(_ context.Context, s *domain.Settings) (*domain.CanonicalSettings, []domain.MigrationItem, error) {
	c := &domain.CanonicalSettings{Languages: make(map[string]domain.CanonicalLanguage)}
	if v, ok := s.Data["editor.formatOnSave"]; ok {
		if b, ok := v.(bool); ok {
			c.Editor.FormatOnSave = b
		}
	}
	return c, []domain.MigrationItem{{SourceKey: "test", Status: domain.FullySupported}}, nil
}

func (n *testNormalizer) FromCanonical(_ context.Context, c *domain.CanonicalSettings) (map[string]any, []domain.MigrationItem, error) {
	out := make(map[string]any)
	if c.Editor.FormatOnSave {
		out["editor.formatOnSave"] = true
	}
	return out, nil, nil
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	store, err := sqlitestore.New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	vsc := &testEditor{
		id: domain.VSCode, name: "VS Code",
		settings:   map[string]any{"editor.tabSize": float64(4), "editor.formatOnSave": true},
		extensions: []domain.Extension{{ID: "golang.go", Version: "0.41.0"}},
	}
	zedE := &testEditor{
		id: domain.Zed, name: "Zed",
		settings: map[string]any{"tab_size": float64(2)},
	}

	editors := map[domain.EditorID]ports.Editor{
		domain.VSCode: vsc,
		domain.Zed:    zedE,
	}
	normalizers := map[domain.EditorID]ports.SettingsNormalizer{
		domain.VSCode: &testNormalizer{},
		domain.Zed:    &testNormalizer{},
	}

	backup := application.NewBackupService(store, editors)
	settings := application.NewSettingsService(editors, backup)
	migration := application.NewMigrationService(editors, normalizers, backup, store)
	docs := application.NewDocsService(store)
	profiles := application.NewProfileService(t.TempDir(), editors)

	return NewServer(settings, backup, migration, docs, profiles, nil, nil)
}

func TestHandleDetectEditors(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleDetectEditors(ctx, &gomcp.CallToolRequest{}, DetectEditorsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	text := extractText(result)
	if !strings.Contains(text, "vscode") && !strings.Contains(text, "VS Code") {
		t.Error("expected VS Code in output")
	}
}

func TestHandleGetSettings(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleGetSettings(ctx, &gomcp.CallToolRequest{}, GetSettingsInput{
		Editor: "vscode", Scope: "user",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "tabSize") && !strings.Contains(text, "tab_size") {
		t.Error("expected tab size in output")
	}
}

func TestHandleGetSettingsUnsupported(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleGetSettings(ctx, &gomcp.CallToolRequest{}, GetSettingsInput{
		Editor: "sublime", Scope: "user",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "Error") && !strings.Contains(text, "unsupported") {
		t.Error("expected error message for unsupported editor")
	}
}

func TestHandleUpdateSettings(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleUpdateSettings(ctx, &gomcp.CallToolRequest{}, UpdateSettingsInput{
		Editor:   "vscode",
		Scope:    "user",
		Settings: map[string]any{"editor.fontSize": float64(16)},
		Strategy: "merge",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "applied") {
		t.Error("expected applied status in output")
	}
}

func TestHandlePlanMigration(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handlePlanMigration(ctx, &gomcp.CallToolRequest{}, PlanMigrationInput{
		Source: "vscode", Target: "zed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "migration") {
		t.Error("expected migration info in output")
	}
}

func TestHandleSearchDocs(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleSearchDocs(ctx, &gomcp.CallToolRequest{}, SearchDocsInput{
		Topic: "settings",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "Settings") {
		t.Error("expected settings docs in output")
	}
}

func TestHandleSearchDocsWithEditor(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleSearchDocs(ctx, &gomcp.CallToolRequest{}, SearchDocsInput{
		Topic:  "settings",
		Editor: "zed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "Zed") && !strings.Contains(text, "zed") {
		t.Error("expected Zed docs in output")
	}
}

func TestHandleFetchDocs(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleFetchDocs(ctx, &gomcp.CallToolRequest{}, FetchDocsInput{
		URL: "https://code.visualstudio.com/docs/getstarted/settings",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "Settings") {
		t.Error("expected settings content in output")
	}
}

func TestHandleFetchDocsNotFound(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleFetchDocs(ctx, &gomcp.CallToolRequest{}, FetchDocsInput{
		URL: "https://example.com/nonexistent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(result)
	if !strings.Contains(text, "not found") {
		t.Error("expected not found message")
	}
}

func TestHandleListProfiles(t *testing.T) {
	srv := newTestServer(t)
	ctx := context.Background()

	result, _, err := srv.handleListProfiles(ctx, &gomcp.CallToolRequest{}, ListProfilesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func extractText(result *gomcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(*gomcp.TextContent); ok {
		return tc.Text
	}
	return ""
}
