package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

// noOutput is used as the second (structured output) return type for tool handlers
// that only return text content.
type noOutput = struct{}

func (s *Server) registerTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "detect_editors",
		Description: "Detect installed code editors and their configuration paths. Returns information about VS Code, Zed, Cursor and other supported editors found on the system.",
	}, s.handleDetectEditors)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_settings",
		Description: "Read current settings for a specific editor. Returns the settings JSON and metadata including the file path.",
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
	Editor string `json:"editor" jsonschema:"description=Editor ID (vscode, zed, cursor),enum=vscode,enum=zed,enum=cursor"`
	Scope  string `json:"scope" jsonschema:"description=Settings scope (user or workspace),enum=user,enum=workspace"`
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
	Editor   string         `json:"editor" jsonschema:"description=Editor ID (vscode, zed, cursor),enum=vscode,enum=zed,enum=cursor"`
	Scope    string         `json:"scope" jsonschema:"description=Settings scope (user or workspace),enum=user,enum=workspace"`
	Settings map[string]any `json:"settings" jsonschema:"description=Settings to apply (key-value pairs)"`
	Strategy string         `json:"merge_strategy" jsonschema:"description=Merge strategy: merge (deep merge default) or replace,enum=merge,enum=replace"`
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
	Editor string `json:"editor" jsonschema:"description=Editor ID (vscode, zed, cursor),enum=vscode,enum=zed,enum=cursor"`
	Label  string `json:"label" jsonschema:"description=Optional label for the backup"`
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
	BackupID string `json:"backup_id" jsonschema:"description=ID of the backup to restore"`
	DryRun   bool   `json:"dry_run" jsonschema:"description=If true preview changes without applying"`
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
	Source string `json:"source" jsonschema:"description=Source editor ID (vscode, zed, cursor),enum=vscode,enum=zed,enum=cursor"`
	Target string `json:"target" jsonschema:"description=Target editor ID (vscode, zed, cursor),enum=vscode,enum=zed,enum=cursor"`
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
	PlanID string `json:"plan_id" jsonschema:"description=ID of the migration plan to execute"`
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

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}
