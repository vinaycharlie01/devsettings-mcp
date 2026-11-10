package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type BackupService struct {
	store   ports.Store
	editors map[domain.EditorID]ports.Editor
}

func NewBackupService(store ports.Store, editors map[domain.EditorID]ports.Editor) *BackupService {
	return &BackupService{store: store, editors: editors}
}

func (s *BackupService) CreateBackup(ctx context.Context, editorID domain.EditorID, label string) (*domain.Backup, error) {
	editor, ok := s.editors[editorID]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", editorID)
	}

	timestamp := time.Now()
	id := fmt.Sprintf("%s-%s", editorID, timestamp.Format("2006-01-02-150405"))

	backup := &domain.Backup{
		ID:        id,
		Editor:    editorID,
		Label:     label,
		Timestamp: timestamp,
	}

	settings, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err == nil && len(settings.Data) > 0 {
		backup.SettingsData = RedactSecrets(settings.Data)
	}

	extensions, err := editor.ListExtensions(ctx)
	if err == nil && len(extensions) > 0 {
		backup.ExtensionsData = extensions
	}

	data, _ := json.Marshal(backup.SettingsData)
	backup.Checksum = checksum(data)

	if err := s.store.SaveBackup(ctx, backup); err != nil {
		return nil, fmt.Errorf("saving backup: %w", err)
	}

	return backup, nil
}

func (s *BackupService) ListBackups(ctx context.Context) ([]domain.Backup, error) {
	return s.store.ListBackups(ctx)
}

func (s *BackupService) RestoreBackup(ctx context.Context, backupID string, dryRun bool) (*domain.SettingsDiff, error) {
	backup, err := s.store.GetBackup(ctx, backupID)
	if err != nil {
		return nil, fmt.Errorf("reading backup: %w", err)
	}

	editor, ok := s.editors[backup.Editor]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", backup.Editor)
	}

	if backup.SettingsData == nil {
		return nil, fmt.Errorf("backup contains no settings data")
	}

	current, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading current settings: %w", err)
	}

	diff := ComputeDiff(current.Data, backup.SettingsData)

	if dryRun {
		return diff, nil
	}

	data, err := json.MarshalIndent(backup.SettingsData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling settings: %w", err)
	}

	if err := editor.WriteSettings(ctx, domain.ScopeUser, data); err != nil {
		return nil, fmt.Errorf("writing restored settings: %w", err)
	}

	return diff, nil
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}`),
	regexp.MustCompile(`(?i)(token|bearer|auth[_-]?token)\s*[:=]\s*["']?[A-Za-z0-9_\-\.]{20,}`),
	regexp.MustCompile(`(?i)(password|passwd|pwd|secret)\s*[:=]\s*["']?.+`),
	regexp.MustCompile(`-----BEGIN .* PRIVATE KEY-----`),
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
}

func RedactSecrets(data map[string]any) map[string]any {
	result := make(map[string]any, len(data))
	for k, v := range data {
		if isSecretKey(k) {
			result[k] = "[REDACTED]"
			continue
		}
		switch val := v.(type) {
		case map[string]any:
			result[k] = RedactSecrets(val)
		case string:
			result[k] = redactSecretValue(val)
		default:
			result[k] = v
		}
	}
	return result
}

func isSecretKey(key string) bool {
	lower := strings.ToLower(key)
	secretKeys := []string{"password", "secret", "token", "apikey", "api_key", "api-key", "private_key", "credentials"}
	for _, s := range secretKeys {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

func redactSecretValue(val string) string {
	for _, pattern := range secretPatterns {
		if pattern.MatchString(val) {
			return "[REDACTED]"
		}
	}
	return val
}

func checksum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
