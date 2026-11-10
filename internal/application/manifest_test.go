package application

import (
	"context"
	"testing"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
	"gopkg.in/yaml.v3"
)

type fakeMCPManager struct {
	config  *domain.MCPConfig
	written []domain.MCPServer
	path    string
}

func (f *fakeMCPManager) ReadMCPConfig(_ context.Context) (*domain.MCPConfig, error) {
	if f.config != nil {
		return f.config, nil
	}
	return &domain.MCPConfig{Editor: domain.VSCode, Path: f.path}, nil
}

func (f *fakeMCPManager) WriteMCPConfig(_ context.Context, servers []domain.MCPServer) error {
	f.written = servers
	return nil
}

func (f *fakeMCPManager) MCPConfigPath() string { return f.path }

func TestParseManifest(t *testing.T) {
	svc := NewManifestService(nil)

	input := `
version: "1"
vars:
  MY_TOKEN: "abc123"
servers:
  - name: test-server
    command: npx
    args: ["-y", "test-mcp"]
    env:
      TOKEN: "${MY_TOKEN}"
  - name: web-server
    url: "http://localhost:3000"
    disabled: [cursor]
    overrides:
      zed:
        command: web-mcp-zed
`

	m, err := svc.ParseManifest([]byte(input))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if m.Version != "1" {
		t.Errorf("version: got %q, want %q", m.Version, "1")
	}
	if len(m.Servers) != 2 {
		t.Fatalf("servers: got %d, want 2", len(m.Servers))
	}
	if m.Servers[0].Name != "test-server" {
		t.Errorf("server[0].name: got %q", m.Servers[0].Name)
	}
	if m.Vars["MY_TOKEN"] != "abc123" {
		t.Errorf("vars.MY_TOKEN: got %q", m.Vars["MY_TOKEN"])
	}
	if len(m.Servers[1].Disabled) != 1 || m.Servers[1].Disabled[0] != domain.Cursor {
		t.Errorf("server[1].disabled: got %v", m.Servers[1].Disabled)
	}
}

func TestValidateManifest(t *testing.T) {
	svc := NewManifestService(nil)

	m := &domain.MCPManifest{
		Servers: []domain.MCPManifestServer{
			{Name: "ok", Command: "cmd"},
			{Name: "", Command: "cmd"},
			{Name: "no-cmd"},
			{Name: "ok", Command: "dup"},
		},
	}

	issues := svc.ValidateManifest(m)
	if len(issues) != 3 {
		t.Errorf("expected 3 issues, got %d: %v", len(issues), issues)
	}
}

func TestSyncToEditor(t *testing.T) {
	vscodeMgr := &fakeMCPManager{path: "/test/vscode/mcp.json"}
	cursorMgr := &fakeMCPManager{path: "/test/cursor/mcp.json"}
	zedMgr := &fakeMCPManager{path: "/test/zed/settings.json"}

	mgrs := map[domain.EditorID]ports.MCPConfigManager{
		domain.VSCode: vscodeMgr,
		domain.Cursor: cursorMgr,
		domain.Zed:    zedMgr,
	}
	svc := NewManifestService(mgrs)

	m := &domain.MCPManifest{
		Version: "1",
		Vars:    map[string]string{"TOKEN": "secret123"},
		Servers: []domain.MCPManifestServer{
			{
				Name:    "main",
				Command: "main-mcp",
				Args:    []string{"--stdio"},
				Env:     map[string]string{"API_KEY": "${TOKEN}"},
			},
			{
				Name:     "vscode-only",
				Command:  "vsc-tool",
				Disabled: []domain.EditorID{domain.Cursor, domain.Zed},
			},
			{
				Name:    "with-override",
				Command: "default-cmd",
				Args:    []string{"--default"},
				Overrides: map[domain.EditorID]domain.MCPServerOverride{
					domain.Zed: {Command: "zed-cmd", Args: []string{"--zed"}},
				},
			},
		},
	}

	ctx := context.Background()

	result, err := svc.SyncToEditor(ctx, m, domain.VSCode)
	if err != nil {
		t.Fatalf("sync vscode: %v", err)
	}
	if result.Servers != 3 {
		t.Errorf("vscode servers: got %d, want 3", result.Servers)
	}
	if result.Skipped != 0 {
		t.Errorf("vscode skipped: got %d, want 0", result.Skipped)
	}
	if vscodeMgr.written[0].Env["API_KEY"] != "secret123" {
		t.Errorf("env not resolved: got %q", vscodeMgr.written[0].Env["API_KEY"])
	}

	result, err = svc.SyncToEditor(ctx, m, domain.Cursor)
	if err != nil {
		t.Fatalf("sync cursor: %v", err)
	}
	if result.Servers != 2 {
		t.Errorf("cursor servers: got %d, want 2 (vscode-only disabled)", result.Servers)
	}
	if result.Skipped != 1 {
		t.Errorf("cursor skipped: got %d, want 1", result.Skipped)
	}

	result, err = svc.SyncToEditor(ctx, m, domain.Zed)
	if err != nil {
		t.Fatalf("sync zed: %v", err)
	}
	if result.Servers != 2 {
		t.Errorf("zed servers: got %d, want 2", result.Servers)
	}
	zedOverridden := zedMgr.written[1]
	if zedOverridden.Command != "zed-cmd" {
		t.Errorf("zed override command: got %q, want %q", zedOverridden.Command, "zed-cmd")
	}
	if len(zedOverridden.Args) != 1 || zedOverridden.Args[0] != "--zed" {
		t.Errorf("zed override args: got %v", zedOverridden.Args)
	}
}

func TestSyncToAll(t *testing.T) {
	vscodeMgr := &fakeMCPManager{path: "/test/vscode/mcp.json"}
	cursorMgr := &fakeMCPManager{path: "/test/cursor/mcp.json"}

	mgrs := map[domain.EditorID]ports.MCPConfigManager{
		domain.VSCode: vscodeMgr,
		domain.Cursor: cursorMgr,
	}
	svc := NewManifestService(mgrs)

	m := &domain.MCPManifest{
		Version: "1",
		Servers: []domain.MCPManifestServer{
			{Name: "srv", Command: "cmd"},
		},
	}

	results, err := svc.SyncToAll(context.Background(), m)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("results: got %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Status != "synced" {
			t.Errorf("editor %s: status %q", r.Editor, r.Status)
		}
	}
}

func TestExportFromEditor(t *testing.T) {
	mgr := &fakeMCPManager{
		path: "/test/vscode/mcp.json",
		config: &domain.MCPConfig{
			Editor: domain.VSCode,
			Servers: []domain.MCPServer{
				{Name: "github", Command: "npx", Args: []string{"-y", "github-mcp"}},
				{Name: "web", URL: "http://localhost:3000"},
			},
		},
	}
	mgrs := map[domain.EditorID]ports.MCPConfigManager{domain.VSCode: mgr}
	svc := NewManifestService(mgrs)

	data, err := svc.ExportFromEditor(context.Background(), domain.VSCode)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	var m domain.MCPManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("re-parse exported YAML: %v", err)
	}
	if len(m.Servers) != 2 {
		t.Errorf("servers: got %d, want 2", len(m.Servers))
	}
	if m.Servers[0].Name != "github" {
		t.Errorf("server[0].name: got %q", m.Servers[0].Name)
	}
	if m.Servers[1].URL != "http://localhost:3000" {
		t.Errorf("server[1].url: got %q", m.Servers[1].URL)
	}
}

func TestExpandVars(t *testing.T) {
	vars := map[string]string{"TOKEN": "abc", "HOST": "example.com"}

	got := expandVars("Bearer ${TOKEN}", vars)
	if got != "Bearer abc" {
		t.Errorf("got %q, want %q", got, "Bearer abc")
	}

	got = expandVars("https://${HOST}/api", vars)
	if got != "https://example.com/api" {
		t.Errorf("got %q, want %q", got, "https://example.com/api")
	}

	got = expandVars("no-vars-here", vars)
	if got != "no-vars-here" {
		t.Errorf("got %q, want %q", got, "no-vars-here")
	}
}

func TestResolveServersEnvFromOS(t *testing.T) {
	t.Setenv("TEST_MCP_KEY", "from-os")

	m := &domain.MCPManifest{
		Servers: []domain.MCPManifestServer{
			{
				Name:    "srv",
				Command: "cmd",
				Env:     map[string]string{"KEY": "${TEST_MCP_KEY}"},
			},
		},
	}

	servers, _ := resolveServers(m, domain.VSCode)
	if servers[0].Env["KEY"] != "from-os" {
		t.Errorf("env from OS: got %q, want %q", servers[0].Env["KEY"], "from-os")
	}
}

func TestGenerateManifestTemplate(t *testing.T) {
	tmpl := GenerateManifestTemplate()
	var m domain.MCPManifest
	if err := yaml.Unmarshal([]byte(tmpl), &m); err != nil {
		t.Fatalf("template is not valid YAML: %v", err)
	}
	if m.Version != "1" {
		t.Errorf("template version: got %q", m.Version)
	}
	if len(m.Servers) == 0 {
		t.Error("template should have at least one example server")
	}
}
