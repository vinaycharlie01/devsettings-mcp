package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type BackupService struct {
	baseDir string
	editors map[domain.EditorID]ports.Editor
}

func NewBackupService(baseDir string, editors map[domain.EditorID]ports.Editor) *BackupService {
	return &BackupService{baseDir: baseDir, editors: editors}
}

func (s *BackupService) CreateBackup(ctx context.Context, editorID domain.EditorID, label string) (*domain.Backup, error) {
	editor, ok := s.editors[editorID]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", editorID)
	}

	timestamp := time.Now()
	id := fmt.Sprintf("%s-%s", editorID, timestamp.Format("2006-01-02-150405"))
	backupDir := filepath.Join(s.baseDir, "backups", id)

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating backup directory: %w", err)
	}

	var files []string
	checksums := make(map[string]string)

	settings, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err == nil && len(settings.Data) > 0 {
		redacted := RedactSecrets(settings.Data)
		data, _ := json.MarshalIndent(redacted, "", "  ")
		settingsFile := "settings.json"
		if err := os.WriteFile(filepath.Join(backupDir, settingsFile), data, 0o644); err != nil {
			return nil, fmt.Errorf("writing settings backup: %w", err)
		}
		files = append(files, settingsFile)
		checksums[settingsFile] = checksum(data)
	}

	extensions, err := editor.ListExtensions(ctx)
	if err == nil && len(extensions) > 0 {
		data, _ := json.MarshalIndent(extensions, "", "  ")
		extFile := "extensions.json"
		if err := os.WriteFile(filepath.Join(backupDir, extFile), data, 0o644); err != nil {
			return nil, fmt.Errorf("writing extensions backup: %w", err)
		}
		files = append(files, extFile)
		checksums[extFile] = checksum(data)
	}

	checksumData, _ := json.MarshalIndent(checksums, "", "  ")
	os.WriteFile(filepath.Join(backupDir, "checksums.json"), checksumData, 0o644)

	manifest := domain.Backup{
		ID:        id,
		Editor:    editorID,
		Label:     label,
		Timestamp: timestamp,
		Dir:       backupDir,
		Files:     files,
	}

	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	os.WriteFile(filepath.Join(backupDir, "manifest.json"), manifestData, 0o644)

	return &manifest, nil
}

func (s *BackupService) ListBackups(ctx context.Context) ([]domain.Backup, error) {
	backupsDir := filepath.Join(s.baseDir, "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading backups directory: %w", err)
	}

	var backups []domain.Backup
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifestPath := filepath.Join(backupsDir, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var backup domain.Backup
		if err := json.Unmarshal(data, &backup); err != nil {
			continue
		}
		backups = append(backups, backup)
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

func (s *BackupService) RestoreBackup(ctx context.Context, backupID string, dryRun bool) (*domain.SettingsDiff, error) {
	backupDir := filepath.Join(s.baseDir, "backups", backupID)
	manifestPath := filepath.Join(backupDir, "manifest.json")

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading backup manifest: %w", err)
	}

	var manifest domain.Backup
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("parsing backup manifest: %w", err)
	}

	editor, ok := s.editors[manifest.Editor]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", manifest.Editor)
	}

	settingsPath := filepath.Join(backupDir, "settings.json")
	backupData, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("reading backup settings: %w", err)
	}

	var backupSettings map[string]any
	if err := json.Unmarshal(backupData, &backupSettings); err != nil {
		return nil, fmt.Errorf("parsing backup settings: %w", err)
	}

	current, err := editor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading current settings: %w", err)
	}

	diff := ComputeDiff(current.Data, backupSettings)

	if dryRun {
		return diff, nil
	}

	data, err := json.MarshalIndent(backupSettings, "", "  ")
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
