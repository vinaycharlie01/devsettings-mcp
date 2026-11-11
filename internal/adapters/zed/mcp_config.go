package zed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func (a *Adapter) MCPConfigPath() string {
	return filepath.Join(a.configDir(), "settings.json")
}

func (a *Adapter) ReadMCPConfig(ctx context.Context) (*domain.MCPConfig, error) {
	settings, err := a.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading Zed settings for MCP config: %w", err)
	}

	config := &domain.MCPConfig{
		Editor: domain.Zed,
		Path:   a.MCPConfigPath(),
	}

	ctxServers, ok := settings.Data["context_servers"].(map[string]any)
	if !ok {
		return config, nil
	}

	for name, val := range ctxServers {
		srv := domain.MCPServer{Name: name}

		if m, ok := val.(map[string]any); ok {
			if cmd, ok := m["command"].(string); ok {
				srv.Command = cmd
			}
			if args, ok := m["args"].([]any); ok {
				for _, a := range args {
					if s, ok := a.(string); ok {
						srv.Args = append(srv.Args, s)
					}
				}
			}
			if env, ok := m["env"].(map[string]any); ok {
				srv.Env = make(map[string]string, len(env))
				for k, v := range env {
					if s, ok := v.(string); ok {
						srv.Env[k] = s
					}
				}
			}
		}

		config.Servers = append(config.Servers, srv)
	}

	return config, nil
}

func (a *Adapter) WriteMCPConfig(ctx context.Context, servers []domain.MCPServer) error {
	settings, err := a.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return fmt.Errorf("reading current Zed settings: %w", err)
	}

	ctxServers := make(map[string]any, len(servers))
	for _, srv := range servers {
		entry := make(map[string]any)
		if srv.Command != "" {
			entry["command"] = srv.Command
			if len(srv.Args) > 0 {
				entry["args"] = srv.Args
			}
		}
		if len(srv.Env) > 0 {
			entry["env"] = srv.Env
		}
		ctxServers[srv.Name] = entry
	}

	settings.Data["context_servers"] = ctxServers

	data, err := json.MarshalIndent(settings.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	path := filepath.Join(a.configDir(), "settings.json")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temp settings: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing settings: %w", err)
	}

	return nil
}
