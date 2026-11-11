package application

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
	"gopkg.in/yaml.v3"
)

type ManifestService struct {
	mcpManagers map[domain.EditorID]ports.MCPConfigManager
}

func NewManifestService(mcpManagers map[domain.EditorID]ports.MCPConfigManager) *ManifestService {
	return &ManifestService{mcpManagers: mcpManagers}
}

func (s *ManifestService) ParseManifest(data []byte) (*domain.MCPManifest, error) {
	var m domain.MCPManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest YAML: %w", err)
	}
	if m.Version == "" {
		m.Version = "1"
	}
	return &m, nil
}

func (s *ManifestService) ValidateManifest(m *domain.MCPManifest) []string {
	var issues []string
	names := make(map[string]bool)
	for i, srv := range m.Servers {
		if srv.Name == "" {
			issues = append(issues, fmt.Sprintf("server[%d]: name is required", i))
		}
		if names[srv.Name] {
			issues = append(issues, fmt.Sprintf("server[%d]: duplicate name %q", i, srv.Name))
		}
		names[srv.Name] = true
		if srv.Command == "" && srv.URL == "" {
			issues = append(issues, fmt.Sprintf("server[%d] %q: needs command or url", i, srv.Name))
		}
	}
	return issues
}

// SyncToEditor resolves the manifest for a specific editor and writes its MCP config.
func (s *ManifestService) SyncToEditor(ctx context.Context, m *domain.MCPManifest, editorID domain.EditorID) (*domain.ManifestSyncResult, error) {
	mgr, ok := s.mcpManagers[editorID]
	if !ok {
		return nil, fmt.Errorf("MCP config not supported for editor: %s", editorID)
	}

	servers, skipped := resolveServers(m, editorID)

	if err := mgr.WriteMCPConfig(ctx, servers); err != nil {
		return nil, fmt.Errorf("writing MCP config for %s: %w", editorID, err)
	}

	return &domain.ManifestSyncResult{
		Editor:  editorID,
		Path:    mgr.MCPConfigPath(),
		Servers: len(servers),
		Skipped: skipped,
		Status:  "synced",
	}, nil
}

// SyncToAll writes the manifest to every known editor.
func (s *ManifestService) SyncToAll(ctx context.Context, m *domain.MCPManifest) ([]domain.ManifestSyncResult, error) {
	var results []domain.ManifestSyncResult
	for editorID := range s.mcpManagers {
		result, err := s.SyncToEditor(ctx, m, editorID)
		if err != nil {
			results = append(results, domain.ManifestSyncResult{
				Editor: editorID,
				Status: fmt.Sprintf("error: %v", err),
			})
			continue
		}
		results = append(results, *result)
	}
	return results, nil
}

// ExportFromEditor reads an editor's MCP config and produces a manifest YAML.
func (s *ManifestService) ExportFromEditor(ctx context.Context, editorID domain.EditorID) ([]byte, error) {
	mgr, ok := s.mcpManagers[editorID]
	if !ok {
		return nil, fmt.Errorf("MCP config not supported for editor: %s", editorID)
	}

	config, err := mgr.ReadMCPConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading MCP config from %s: %w", editorID, err)
	}

	manifest := domain.MCPManifest{Version: "1"}
	for _, srv := range config.Servers {
		manifest.Servers = append(manifest.Servers, domain.MCPManifestServer{
			Name:      srv.Name,
			Command:   srv.Command,
			Args:      srv.Args,
			URL:       srv.URL,
			Transport: srv.Transport,
			Env:       srv.Env,
		})
	}

	return yaml.Marshal(&manifest)
}

func resolveServers(m *domain.MCPManifest, editorID domain.EditorID) ([]domain.MCPServer, int) {
	var servers []domain.MCPServer
	skipped := 0

	for _, ms := range m.Servers {
		if isDisabled(ms.Disabled, editorID) {
			skipped++
			continue
		}

		srv := domain.MCPServer{
			Name:      ms.Name,
			Command:   ms.Command,
			Args:      copyStrings(ms.Args),
			URL:       ms.URL,
			Transport: ms.Transport,
			Env:       resolveEnv(ms.Env, m.Vars),
		}

		if ov, ok := ms.Overrides[editorID]; ok {
			if ov.Command != "" {
				srv.Command = ov.Command
			}
			if ov.Args != nil {
				srv.Args = copyStrings(ov.Args)
			}
			if ov.URL != "" {
				srv.URL = ov.URL
			}
			if ov.Env != nil {
				srv.Env = resolveEnv(ov.Env, m.Vars)
			}
		}

		servers = append(servers, srv)
	}

	return servers, skipped
}

func isDisabled(disabled []domain.EditorID, target domain.EditorID) bool {
	for _, d := range disabled {
		if d == target {
			return true
		}
	}
	return false
}

func resolveEnv(env map[string]string, vars map[string]string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	resolved := make(map[string]string, len(env))
	for k, v := range env {
		resolved[k] = expandVars(v, vars)
	}
	return resolved
}

func expandVars(s string, vars map[string]string) string {
	return os.Expand(s, func(key string) string {
		if v, ok := vars[key]; ok {
			return v
		}
		return os.Getenv(key)
	})
}

func copyStrings(s []string) []string {
	if s == nil {
		return nil
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}

// GenerateManifestTemplate produces a starter YAML manifest.
func GenerateManifestTemplate() string {
	return strings.TrimSpace(`
version: "1"

# Variables — referenced as ${VAR_NAME} in env values
vars:
  GITHUB_TOKEN: ""

servers:
  - name: example-server
    command: npx
    args: ["-y", "example-mcp-server"]
    env:
      API_TOKEN: "${GITHUB_TOKEN}"

    # Disable for specific editors
    # disabled: [cursor]

    # Per-editor overrides
    # overrides:
    #   zed:
    #     command: example-mcp-zed
    #     args: []
`) + "\n"
}
