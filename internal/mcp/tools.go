package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type noOutput = struct{}

func (s *Server) registerTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "detect_editors",
		Description: "Detect installed code editors and their configuration paths. Returns information about VS Code, Zed, Cursor and other supported editors found on the system.",
	}, s.handleDetectEditors)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_settings",
		Description: "Read current settings for a specific editor. Editor: vscode, zed, or cursor. Scope: user or workspace.",
	}, s.handleGetSettings)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "update_settings",
		Description: "Update settings for a specific editor. Performs a deep merge by default, creates a backup before modifying, and returns a diff of changes.",
	}, s.handleUpdateSettings)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "create_backup",
		Description: "Create a backup of an editor's configuration. Backs up settings, extensions list, and other configuration files. Secrets are automatically redacted.",
	}, s.handleCreateBackup)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "restore_backup",
		Description: "Restore a previous backup. Supports dry-run mode to preview changes before applying.",
	}, s.handleRestoreBackup)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "plan_migration",
		Description: "Generate a migration plan between two editors. Analyzes settings compatibility, identifies supported and unsupported features, and creates a detailed report.",
	}, s.handlePlanMigration)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "execute_migration",
		Description: "Execute a previously generated migration plan. Requires prior approval. Creates backups of both source and target before applying changes.",
	}, s.handleExecuteMigration)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "search_docs",
		Description: "Search editor and language documentation for a topic. Returns matching documentation sections with URLs. Topic examples: settings, keybindings, go, python, extensions. Optional editor filter: vscode, zed, cursor.",
	}, s.handleSearchDocs)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "fetch_docs",
		Description: "Retrieve a specific documentation page by URL. Optionally filter by section heading. Returns the documentation content.",
	}, s.handleFetchDocs)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_profiles",
		Description: "List available development profiles. Profiles are pre-defined configurations for specific development workflows (e.g., Go backend, web frontend, Rust).",
	}, s.handleListProfiles)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "load_profile",
		Description: "Load and apply a development profile to an editor. Merges profile settings with the editor's current configuration. Editor: vscode, zed, or cursor.",
	}, s.handleLoadProfile)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_ai_settings",
		Description: "Read AI-specific settings for an editor. Returns the canonical AI model (inline completion, chat, model, provider) plus raw AI keys. Shows cross-editor mapping table. Editor: vscode, zed, or cursor.",
	}, s.handleGetAISettings)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "update_ai_settings",
		Description: "Update AI settings for an editor using the canonical model. Fields: inline_completion (bool), chat_enabled (bool), model (string), provider (string), composer_enabled (bool, Cursor only), prediction_enabled (bool, Cursor only). Editor: vscode, zed, or cursor.",
	}, s.handleUpdateAISettings)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_mcp_config",
		Description: "Read MCP server configurations for an editor. Returns the list of configured MCP servers with their command, args, URL, transport, and env. Editor: vscode, zed, or cursor.",
	}, s.handleGetMCPConfig)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "update_mcp_config",
		Description: "Add or update MCP server configurations for an editor. Provide a list of servers with name, command, args, url, and env. Replaces the full MCP config. Editor: vscode, zed, or cursor.",
	}, s.handleUpdateMCPConfig)
}

type DetectEditorsInput struct{}

func (s *Server) handleDetectEditors(ctx context.Context, _ *mcp.CallToolRequest, _ DetectEditorsInput) (*mcp.CallToolResult, noOutput, error) {
	editors, err := s.settings.DetectEditors(ctx)
	if err != nil {
		return textResult(fmt.Sprintf("Error detecting editors: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(editors, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type GetSettingsInput struct {
	Editor string `json:"editor"`
	Scope  string `json:"scope"`
}

func (s *Server) handleGetSettings(ctx context.Context, _ *mcp.CallToolRequest, input GetSettingsInput) (*mcp.CallToolResult, noOutput, error) {
	scope := domain.ScopeUser
	if input.Scope == "workspace" {
		scope = domain.ScopeWorkspace
	}

	settings, err := s.settings.GetSettings(ctx, domain.EditorID(input.Editor), scope)
	if err != nil {
		return textResult(fmt.Sprintf("Error reading settings: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(settings, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type UpdateSettingsInput struct {
	Editor   string         `json:"editor"`
	Scope    string         `json:"scope"`
	Settings map[string]any `json:"settings"`
	Strategy string         `json:"merge_strategy"`
}

func (s *Server) handleUpdateSettings(ctx context.Context, _ *mcp.CallToolRequest, input UpdateSettingsInput) (*mcp.CallToolResult, noOutput, error) {
	scope := domain.ScopeUser
	if input.Scope == "workspace" {
		scope = domain.ScopeWorkspace
	}

	strategy := domain.MergeDeep
	if input.Strategy == "replace" {
		strategy = domain.MergeReplace
	}

	update := domain.SettingsUpdate{
		Editor:   domain.EditorID(input.Editor),
		Scope:    scope,
		Settings: input.Settings,
		Strategy: strategy,
	}

	diff, backupID, err := s.settings.UpdateSettings(ctx, update)
	if err != nil {
		return textResult(fmt.Sprintf("Error updating settings: %v", err)), noOutput{}, nil
	}

	result := map[string]any{
		"diff":      diff,
		"backup_id": backupID,
		"status":    "applied",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type CreateBackupInput struct {
	Editor string `json:"editor"`
	Label  string `json:"label"`
}

func (s *Server) handleCreateBackup(ctx context.Context, _ *mcp.CallToolRequest, input CreateBackupInput) (*mcp.CallToolResult, noOutput, error) {
	backup, err := s.backup.CreateBackup(ctx, domain.EditorID(input.Editor), input.Label)
	if err != nil {
		return textResult(fmt.Sprintf("Error creating backup: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(backup, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type RestoreBackupInput struct {
	BackupID string `json:"backup_id"`
	DryRun   bool   `json:"dry_run"`
}

func (s *Server) handleRestoreBackup(ctx context.Context, _ *mcp.CallToolRequest, input RestoreBackupInput) (*mcp.CallToolResult, noOutput, error) {
	diff, err := s.backup.RestoreBackup(ctx, input.BackupID, input.DryRun)
	if err != nil {
		return textResult(fmt.Sprintf("Error restoring backup: %v", err)), noOutput{}, nil
	}

	result := map[string]any{
		"diff":    diff,
		"dry_run": input.DryRun,
	}
	if input.DryRun {
		result["status"] = "preview"
	} else {
		result["status"] = "restored"
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type PlanMigrationInput struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func (s *Server) handlePlanMigration(ctx context.Context, _ *mcp.CallToolRequest, input PlanMigrationInput) (*mcp.CallToolResult, noOutput, error) {
	plan, err := s.migration.PlanMigration(ctx, domain.EditorID(input.Source), domain.EditorID(input.Target))
	if err != nil {
		return textResult(fmt.Sprintf("Error planning migration: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(plan, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type ExecuteMigrationInput struct {
	PlanID string `json:"plan_id"`
}

func (s *Server) handleExecuteMigration(ctx context.Context, _ *mcp.CallToolRequest, input ExecuteMigrationInput) (*mcp.CallToolResult, noOutput, error) {
	err := s.migration.ExecuteMigration(ctx, input.PlanID)
	if err != nil {
		return textResult(fmt.Sprintf("Error executing migration: %v", err)), noOutput{}, nil
	}

	result := map[string]any{
		"plan_id": input.PlanID,
		"status":  "completed",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type SearchDocsInput struct {
	Topic  string `json:"topic"`
	Editor string `json:"editor"`
}

func (s *Server) handleSearchDocs(ctx context.Context, _ *mcp.CallToolRequest, input SearchDocsInput) (*mcp.CallToolResult, noOutput, error) {
	if s.docs == nil {
		return textResult("Documentation service not available"), noOutput{}, nil
	}

	result, err := s.docs.Search(ctx, input.Topic, input.Editor)
	if err != nil {
		return textResult(fmt.Sprintf("Error searching docs: %v", err)), noOutput{}, nil
	}

	if len(result.Entries) == 0 {
		return textResult(fmt.Sprintf("No documentation found for topic: %s", input.Topic)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type FetchDocsInput struct {
	URL     string `json:"url"`
	Section string `json:"section"`
}

func (s *Server) handleFetchDocs(ctx context.Context, _ *mcp.CallToolRequest, input FetchDocsInput) (*mcp.CallToolResult, noOutput, error) {
	if s.docs == nil {
		return textResult("Documentation service not available"), noOutput{}, nil
	}

	entry, err := s.docs.Fetch(ctx, input.URL, input.Section)
	if err != nil {
		return textResult(fmt.Sprintf("Error fetching docs: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type ListProfilesInput struct{}

func (s *Server) handleListProfiles(ctx context.Context, _ *mcp.CallToolRequest, _ ListProfilesInput) (*mcp.CallToolResult, noOutput, error) {
	if s.profiles == nil {
		return textResult("Profile service not available"), noOutput{}, nil
	}

	profiles, err := s.profiles.ListProfiles(ctx)
	if err != nil {
		return textResult(fmt.Sprintf("Error listing profiles: %v", err)), noOutput{}, nil
	}

	if len(profiles) == 0 {
		return textResult("No profiles found. Add YAML profile files to the profiles directory."), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(profiles, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type LoadProfileInput struct {
	Name   string `json:"name"`
	Editor string `json:"editor"`
}

func (s *Server) handleLoadProfile(ctx context.Context, _ *mcp.CallToolRequest, input LoadProfileInput) (*mcp.CallToolResult, noOutput, error) {
	if s.profiles == nil {
		return textResult("Profile service not available"), noOutput{}, nil
	}

	applied, err := s.profiles.ApplyProfile(ctx, input.Name, domain.EditorID(input.Editor))
	if err != nil {
		return textResult(fmt.Sprintf("Error applying profile: %v", err)), noOutput{}, nil
	}

	result := map[string]any{
		"profile":          input.Name,
		"editor":           input.Editor,
		"applied_settings": applied,
		"status":           "applied",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type GetAISettingsInput struct {
	Editor string `json:"editor"`
}

func (s *Server) handleGetAISettings(ctx context.Context, _ *mcp.CallToolRequest, input GetAISettingsInput) (*mcp.CallToolResult, noOutput, error) {
	if s.ai == nil {
		return textResult("AI settings service not available"), noOutput{}, nil
	}

	report, err := s.ai.GetAISettings(ctx, domain.EditorID(input.Editor))
	if err != nil {
		return textResult(fmt.Sprintf("Error reading AI settings: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(report, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type UpdateAISettingsInput struct {
	Editor            string `json:"editor"`
	InlineCompletion  bool   `json:"inline_completion"`
	ChatEnabled       bool   `json:"chat_enabled"`
	Model             string `json:"model"`
	Provider          string `json:"provider"`
	ComposerEnabled   bool   `json:"composer_enabled"`
	PredictionEnabled bool   `json:"prediction_enabled"`
}

func (s *Server) handleUpdateAISettings(ctx context.Context, _ *mcp.CallToolRequest, input UpdateAISettingsInput) (*mcp.CallToolResult, noOutput, error) {
	if s.ai == nil {
		return textResult("AI settings service not available"), noOutput{}, nil
	}

	ai := domain.CanonicalAI{
		InlineCompletion:  input.InlineCompletion,
		ChatEnabled:       input.ChatEnabled,
		Model:             input.Model,
		Provider:          input.Provider,
		ComposerEnabled:   input.ComposerEnabled,
		PredictionEnabled: input.PredictionEnabled,
	}

	applied, err := s.ai.UpdateAISettings(ctx, domain.EditorID(input.Editor), ai)
	if err != nil {
		return textResult(fmt.Sprintf("Error updating AI settings: %v", err)), noOutput{}, nil
	}

	result := map[string]any{
		"editor":           input.Editor,
		"applied_settings": applied,
		"status":           "applied",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type GetMCPConfigInput struct {
	Editor string `json:"editor"`
}

func (s *Server) handleGetMCPConfig(ctx context.Context, _ *mcp.CallToolRequest, input GetMCPConfigInput) (*mcp.CallToolResult, noOutput, error) {
	if s.ai == nil {
		return textResult("AI settings service not available"), noOutput{}, nil
	}

	config, err := s.ai.GetMCPConfig(ctx, domain.EditorID(input.Editor))
	if err != nil {
		return textResult(fmt.Sprintf("Error reading MCP config: %v", err)), noOutput{}, nil
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

type UpdateMCPConfigInput struct {
	Editor  string             `json:"editor"`
	Servers []domain.MCPServer `json:"servers"`
}

func (s *Server) handleUpdateMCPConfig(ctx context.Context, _ *mcp.CallToolRequest, input UpdateMCPConfigInput) (*mcp.CallToolResult, noOutput, error) {
	if s.ai == nil {
		return textResult("AI settings service not available"), noOutput{}, nil
	}

	if err := s.ai.UpdateMCPConfig(ctx, domain.EditorID(input.Editor), input.Servers); err != nil {
		return textResult(fmt.Sprintf("Error updating MCP config: %v", err)), noOutput{}, nil
	}

	result := map[string]any{
		"editor":  input.Editor,
		"servers": len(input.Servers),
		"status":  "applied",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return textResult(string(data)), noOutput{}, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}
