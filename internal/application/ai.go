package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type AISettingsService struct {
	editors     map[domain.EditorID]ports.Editor
	normalizers map[domain.EditorID]ports.SettingsNormalizer
	mcpManagers map[domain.EditorID]ports.MCPConfigManager
}

func NewAISettingsService(
	editors map[domain.EditorID]ports.Editor,
	normalizers map[domain.EditorID]ports.SettingsNormalizer,
	mcpManagers map[domain.EditorID]ports.MCPConfigManager,
) *AISettingsService {
	return &AISettingsService{
		editors:     editors,
		normalizers: normalizers,
		mcpManagers: mcpManagers,
	}
}

type AISettingsReport struct {
	Editor   domain.EditorID            `json:"editor"`
	AI       domain.CanonicalAI         `json:"ai"`
	Mappings []domain.AISettingsMapping `json:"mappings"`
	RawKeys  map[string]any             `json:"raw_keys,omitempty"`
}

func (s *AISettingsService) GetAISettings(ctx context.Context, editorID domain.EditorID) (*AISettingsReport, error) {
	editor, ok := s.editors[editorID]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", editorID)
	}

	norm, ok := s.normalizers[editorID]
	if !ok {
		return nil, fmt.Errorf("no normalizer for editor: %s", editorID)
	}

	settings, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	canonical, _, err := norm.ToCanonical(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("normalizing settings: %w", err)
	}

	rawAI := extractAIKeys(editorID, settings.Data)

	return &AISettingsReport{
		Editor:   editorID,
		AI:       canonical.AI,
		Mappings: domain.AISettingsMappings(),
		RawKeys:  rawAI,
	}, nil
}

func (s *AISettingsService) UpdateAISettings(ctx context.Context, editorID domain.EditorID, ai domain.CanonicalAI) (map[string]any, error) {
	editor, ok := s.editors[editorID]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", editorID)
	}

	settings := aiToEditorSettings(editorID, ai)

	current, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading current settings: %w", err)
	}

	merged := DeepMerge(current.Data, settings)

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling settings: %w", err)
	}

	if err := editor.WriteSettings(ctx, domain.ScopeUser, data); err != nil {
		return nil, fmt.Errorf("writing settings: %w", err)
	}

	return settings, nil
}

func (s *AISettingsService) GetMCPConfig(ctx context.Context, editorID domain.EditorID) (*domain.MCPConfig, error) {
	mgr, ok := s.mcpManagers[editorID]
	if !ok {
		return nil, fmt.Errorf("MCP config not supported for editor: %s", editorID)
	}

	return mgr.ReadMCPConfig(ctx)
}

func (s *AISettingsService) UpdateMCPConfig(ctx context.Context, editorID domain.EditorID, servers []domain.MCPServer) error {
	mgr, ok := s.mcpManagers[editorID]
	if !ok {
		return fmt.Errorf("MCP config not supported for editor: %s", editorID)
	}

	return mgr.WriteMCPConfig(ctx, servers)
}

func extractAIKeys(editorID domain.EditorID, data map[string]any) map[string]any {
	result := make(map[string]any)
	var prefixes []string

	switch editorID {
	case domain.VSCode:
		prefixes = []string{"github.copilot"}
	case domain.Cursor:
		prefixes = []string{"cursor.tab", "cursor.chat", "cursor.ai", "cursor.composer", "cursor.prediction", "cursor.copilot"}
	case domain.Zed:
		prefixes = []string{"assistant", "features.inline_completion"}
	}

	for key, val := range data {
		for _, prefix := range prefixes {
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				result[key] = val
				break
			}
		}
	}

	return result
}

func aiToEditorSettings(editorID domain.EditorID, ai domain.CanonicalAI) map[string]any {
	settings := make(map[string]any)

	switch editorID {
	case domain.VSCode:
		if ai.InlineCompletion {
			settings["github.copilot.enable"] = true
		}
		if ai.ChatEnabled {
			settings["github.copilot.chat.enabled"] = true
		}
		if ai.Model != "" {
			settings["github.copilot.advanced.model"] = ai.Model
		}

	case domain.Cursor:
		if ai.InlineCompletion {
			settings["cursor.tab.enabled"] = true
		}
		if ai.ChatEnabled {
			settings["cursor.chat.enabled"] = true
		}
		if ai.Model != "" {
			settings["cursor.ai.model"] = ai.Model
		}
		if ai.ComposerEnabled {
			settings["cursor.composer.enabled"] = true
		}
		if ai.PredictionEnabled {
			settings["cursor.prediction.enabled"] = true
		}

	case domain.Zed:
		if ai.InlineCompletion {
			provider := ai.Provider
			if provider == "" {
				provider = "copilot"
			}
			settings["features.inline_completion_provider"] = provider
		}
		if ai.ChatEnabled {
			settings["assistant.enabled"] = true
		}
		if ai.Model != "" {
			provider := ai.Provider
			if provider == "" {
				provider = "zed.dev"
			}
			settings["assistant.default_model"] = map[string]any{
				"provider": provider,
				"model":    ai.Model,
			}
		}
	}

	return settings
}
