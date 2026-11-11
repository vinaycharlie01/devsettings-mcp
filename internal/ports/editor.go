package ports

import (
	"context"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type Editor interface {
	ID() domain.EditorID
	DisplayName() string
	Detect(ctx context.Context) (*domain.EditorInfo, error)
	ReadSettings(ctx context.Context, scope domain.SettingsScope) (*domain.Settings, error)
	WriteSettings(ctx context.Context, scope domain.SettingsScope, data []byte) error
	SettingsPath(scope domain.SettingsScope) string
	ListExtensions(ctx context.Context) ([]domain.Extension, error)
}

type ExtensionInstaller interface {
	InstallExtension(ctx context.Context, id string) error
	UninstallExtension(ctx context.Context, id string) error
}

type SettingsNormalizer interface {
	ToCanonical(ctx context.Context, settings *domain.Settings) (*domain.CanonicalSettings, []domain.MigrationItem, error)
	FromCanonical(ctx context.Context, canonical *domain.CanonicalSettings) (map[string]any, []domain.MigrationItem, error)
}
