package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

func TestStoreBackupRoundTrip(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	backup := &domain.Backup{
		ID:        "vscode-2024-01-01-120000",
		Editor:    domain.VSCode,
		Label:     "test backup",
		Timestamp: time.Now().Truncate(time.Second),
		SettingsData: map[string]any{
			"editor.tabSize":      float64(4),
			"editor.formatOnSave": true,
		},
		ExtensionsData: []domain.Extension{
			{ID: "golang.go", Version: "0.41.0"},
		},
		Checksum: "abc123",
	}

	if err := store.SaveBackup(ctx, backup); err != nil {
		t.Fatalf("saving backup: %v", err)
	}

	got, err := store.GetBackup(ctx, backup.ID)
	if err != nil {
		t.Fatalf("getting backup: %v", err)
	}

	if got.ID != backup.ID {
		t.Errorf("ID: got %s, want %s", got.ID, backup.ID)
	}
	if got.Editor != backup.Editor {
		t.Errorf("Editor: got %s, want %s", got.Editor, backup.Editor)
	}
	if got.Label != backup.Label {
		t.Errorf("Label: got %s, want %s", got.Label, backup.Label)
	}
	if got.Checksum != backup.Checksum {
		t.Errorf("Checksum: got %s, want %s", got.Checksum, backup.Checksum)
	}

	tabSize, ok := got.SettingsData["editor.tabSize"]
	if !ok {
		t.Fatal("missing editor.tabSize in restored settings")
	}
	if tabSize != float64(4) {
		t.Errorf("editor.tabSize: got %v, want 4", tabSize)
	}

	if len(got.ExtensionsData) != 1 {
		t.Fatalf("extensions: got %d, want 1", len(got.ExtensionsData))
	}
	if got.ExtensionsData[0].ID != "golang.go" {
		t.Errorf("extension ID: got %s, want golang.go", got.ExtensionsData[0].ID)
	}
}

func TestStoreListBackups(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	for i, id := range []string{"b1", "b2", "b3"} {
		store.SaveBackup(ctx, &domain.Backup{
			ID:        id,
			Editor:    domain.VSCode,
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		})
	}

	backups, err := store.ListBackups(ctx)
	if err != nil {
		t.Fatalf("listing backups: %v", err)
	}
	if len(backups) != 3 {
		t.Fatalf("got %d backups, want 3", len(backups))
	}
	if backups[0].ID != "b3" {
		t.Errorf("first backup: got %s, want b3 (most recent)", backups[0].ID)
	}
}

func TestStoreBackupNotFound(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	_, err = store.GetBackup(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backup")
	}
}

func TestStoreMigrationPlanRoundTrip(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	plan := &domain.MigrationPlan{
		ID:     "migration-vscode-zed-2024-01-01",
		Source: domain.VSCode,
		Target: domain.Zed,
		Items: []domain.MigrationItem{
			{SourceKey: "editor.formatOnSave", TargetKey: "format_on_save", Status: domain.SupportedTransformed},
			{SourceKey: "editor.tabSize", TargetKey: "tab_size", Status: domain.FullySupported},
		},
		BackupID: "backup-123",
	}

	if err := store.SaveMigrationPlan(ctx, plan); err != nil {
		t.Fatalf("saving plan: %v", err)
	}

	got, err := store.GetMigrationPlan(ctx, plan.ID)
	if err != nil {
		t.Fatalf("getting plan: %v", err)
	}

	if got.Source != domain.VSCode {
		t.Errorf("Source: got %s, want vscode", got.Source)
	}
	if got.Target != domain.Zed {
		t.Errorf("Target: got %s, want zed", got.Target)
	}
	if len(got.Items) != 2 {
		t.Fatalf("Items: got %d, want 2", len(got.Items))
	}
	if got.Items[0].Status != domain.SupportedTransformed {
		t.Errorf("Item[0].Status: got %s, want %s", got.Items[0].Status, domain.SupportedTransformed)
	}
}

func TestStoreMigrationPlanNotFound(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	_, err = store.GetMigrationPlan(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plan")
	}
}

func TestStoreDocsCacheRoundTrip(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	entry := &domain.DocEntry{
		Title:  "Test Doc",
		URL:    "https://example.com/docs",
		Editor: domain.VSCode,
		Topic:  "testing",
		Sections: []domain.DocSection{
			{Heading: "Overview", Content: "Test content"},
		},
	}

	if err := store.CacheDoc(ctx, entry); err != nil {
		t.Fatalf("caching doc: %v", err)
	}

	got, err := store.GetCachedDoc(ctx, entry.URL)
	if err != nil {
		t.Fatalf("getting cached doc: %v", err)
	}

	if got.Title != entry.Title {
		t.Errorf("Title: got %s, want %s", got.Title, entry.Title)
	}
	if got.Topic != entry.Topic {
		t.Errorf("Topic: got %s, want %s", got.Topic, entry.Topic)
	}
	if len(got.Sections) != 1 {
		t.Fatalf("Sections: got %d, want 1", len(got.Sections))
	}
}

func TestStoreDocsCacheNotFound(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	_, err = store.GetCachedDoc(context.Background(), "https://example.com/nonexistent")
	if err == nil {
		t.Error("expected error for uncached doc")
	}
}

func TestStoreDocsCacheUpdate(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	url := "https://example.com/docs"

	store.CacheDoc(ctx, &domain.DocEntry{
		Title: "Version 1", URL: url, Topic: "test",
	})

	store.CacheDoc(ctx, &domain.DocEntry{
		Title: "Version 2", URL: url, Topic: "test",
	})

	got, err := store.GetCachedDoc(ctx, url)
	if err != nil {
		t.Fatalf("getting cached doc: %v", err)
	}
	if got.Title != "Version 2" {
		t.Errorf("Title: got %s, want Version 2 (should be updated)", got.Title)
	}
}
