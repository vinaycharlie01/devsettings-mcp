package ports

import (
	"context"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type Store interface {
	SaveBackup(ctx context.Context, backup *domain.Backup) error
	GetBackup(ctx context.Context, id string) (*domain.Backup, error)
	ListBackups(ctx context.Context) ([]domain.Backup, error)

	SaveMigrationPlan(ctx context.Context, plan *domain.MigrationPlan) error
	GetMigrationPlan(ctx context.Context, id string) (*domain.MigrationPlan, error)

	CacheDoc(ctx context.Context, entry *domain.DocEntry) error
	GetCachedDoc(ctx context.Context, url string) (*domain.DocEntry, error)

	Close() error
}
