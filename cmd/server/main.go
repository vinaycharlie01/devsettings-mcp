package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/vinaycharlie01/devsettings-mcp/internal/adapters/cursor"
	"github.com/vinaycharlie01/devsettings-mcp/internal/adapters/vscode"
	"github.com/vinaycharlie01/devsettings-mcp/internal/adapters/zed"
	"github.com/vinaycharlie01/devsettings-mcp/internal/application"
	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	mcpserver "github.com/vinaycharlie01/devsettings-mcp/internal/mcp"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting home directory: %v\n", err)
		os.Exit(1)
	}

	workspaceDir, _ := os.Getwd()

	// Editor adapters
	vscodeAdapter := vscode.New(homeDir, workspaceDir)
	zedAdapter := zed.New(homeDir, workspaceDir)
	cursorAdapter := cursor.New(homeDir, workspaceDir)

	editors := map[domain.EditorID]ports.Editor{
		domain.VSCode: vscodeAdapter,
		domain.Zed:    zedAdapter,
		domain.Cursor: cursorAdapter,
	}

	normalizers := map[domain.EditorID]ports.SettingsNormalizer{
		domain.VSCode: vscode.NewNormalizer(),
		domain.Zed:    zed.NewNormalizer(),
		domain.Cursor: cursor.NewNormalizer(),
	}

	// Data directory for backups, profiles, and docs cache
	dataDir := filepath.Join(homeDir, ".devsettings")
	profileDir := filepath.Join(workspaceDir, "profiles")

	// Application services
	backupService := application.NewBackupService(dataDir, editors)
	settingsService := application.NewSettingsService(editors, backupService)
	migrationService := application.NewMigrationService(editors, normalizers, backupService)
	docsService := application.NewDocsService(dataDir)
	profileService := application.NewProfileService(profileDir, editors)

	// MCP server
	server := mcpserver.NewServer(settingsService, backupService, migrationService, docsService, profileService)

	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
