package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/vinaycharlie01/devsettings-mcp/internal/application"
)

type Server struct {
	server    *mcp.Server
	settings  *application.SettingsService
	backup    *application.BackupService
	migration *application.MigrationService
	docs      *application.DocsService
	profiles  *application.ProfileService
}

func NewServer(
	settings *application.SettingsService,
	backup *application.BackupService,
	migration *application.MigrationService,
	docs *application.DocsService,
	profiles *application.ProfileService,
) *Server {
	s := &Server{
		settings:  settings,
		backup:    backup,
		migration: migration,
		docs:      docs,
		profiles:  profiles,
	}

	s.server = mcp.NewServer(
		&mcp.Implementation{
			Name:    "devsettings-mcp",
			Version: "0.1.0",
		},
		nil,
	)

	s.registerTools()
	return s
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}
