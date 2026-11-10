package vscode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func (a *Adapter) MCPConfigPath() string {
	if a.workspaceDir != "" {
		return filepath.Join(a.workspaceDir, ".vscode", "mcp.json")
	}
	return filepath.Join(a.configDir(), "mcp.json")
}

func (a *Adapter) ReadMCPConfig(_ context.Context) (*domain.MCPConfig, error) {
	path := a.MCPConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &domain.MCPConfig{
				Editor: domain.VSCode, Path: path,
			}, nil
		}
		return nil, fmt.Errorf("reading MCP config: %w", err)
	}

	var raw struct {
		Servers map[string]json.RawMessage `json:"servers"`
		Inputs  []any                      `json:"inputs,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing MCP config: %w", err)
	}

	config := &domain.MCPConfig{
		Editor: domain.VSCode,
		Path:   path,
	}

	for name, serverJSON := range raw.Servers {
		srv := parseMCPServer(name, serverJSON)
		config.Servers = append(config.Servers, srv)
	}

	return config, nil
}

func (a *Adapter) WriteMCPConfig(_ context.Context, servers []domain.MCPServer) error {
	path := a.MCPConfigPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	serversMap := make(map[string]any, len(servers))
	for _, srv := range servers {
		entry := make(map[string]any)
		if srv.Command != "" {
			entry["command"] = srv.Command
			if len(srv.Args) > 0 {
				entry["args"] = srv.Args
			}
			entry["type"] = "stdio"
		} else if srv.URL != "" {
			entry["url"] = srv.URL
			entry["type"] = "sse"
		}
		if len(srv.Env) > 0 {
			entry["env"] = srv.Env
		}
		serversMap[srv.Name] = entry
	}

	out := map[string]any{"servers": serversMap}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling MCP config: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temp MCP config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing MCP config: %w", err)
	}

	return nil
}

func parseMCPServer(name string, raw json.RawMessage) domain.MCPServer {
	srv := domain.MCPServer{Name: name}

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return srv
	}

	if cmd, ok := parsed["command"].(string); ok {
		srv.Command = cmd
	}
	if args, ok := parsed["args"].([]any); ok {
		for _, a := range args {
			if s, ok := a.(string); ok {
				srv.Args = append(srv.Args, s)
			}
		}
	}
	if url, ok := parsed["url"].(string); ok {
		srv.URL = url
	}
	if t, ok := parsed["type"].(string); ok {
		srv.Transport = t
	}
	if env, ok := parsed["env"].(map[string]any); ok {
		srv.Env = make(map[string]string, len(env))
		for k, v := range env {
			if s, ok := v.(string); ok {
				srv.Env[k] = s
			}
		}
	}

	return srv
}
