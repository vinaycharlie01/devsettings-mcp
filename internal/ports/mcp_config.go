package ports

import (
	"context"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type MCPConfigManager interface {
	ReadMCPConfig(ctx context.Context) (*domain.MCPConfig, error)
	WriteMCPConfig(ctx context.Context, servers []domain.MCPServer) error
	MCPConfigPath() string
}
