package ports

import (
	"context"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type DocFetcher interface {
	Fetch(ctx context.Context, url string) (*domain.DocEntry, error)
}
