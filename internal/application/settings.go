package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type SettingsService struct {
	editors map[domain.EditorID]ports.Editor
	backup  *BackupService
}

func NewSettingsService(editors map[domain.EditorID]ports.Editor, backup *BackupService) *SettingsService {
	return &SettingsService{editors: editors, backup: backup}
}

func (s *SettingsService) DetectEditors(ctx context.Context) ([]domain.EditorInfo, error) {
	var results []domain.EditorInfo
	for _, editor := range s.editors {
		info, err := editor.Detect(ctx)
		if err != nil {
			continue
		}
		results = append(results, *info)
	}
	return results, nil
}

func (s *SettingsService) GetSettings(ctx context.Context, editorID domain.EditorID, scope domain.SettingsScope) (*domain.Settings, error) {
	editor, ok := s.editors[editorID]
	if !ok {
		return nil, fmt.Errorf("unsupported editor: %s", editorID)
	}
	return editor.ReadSettings(ctx, scope)
}

func (s *SettingsService) UpdateSettings(ctx context.Context, update domain.SettingsUpdate) (*domain.SettingsDiff, string, error) {
	editor, ok := s.editors[update.Editor]
	if !ok {
		return nil, "", fmt.Errorf("unsupported editor: %s", update.Editor)
	}

	current, err := editor.ReadSettings(ctx, update.Scope)
	if err != nil {
		return nil, "", fmt.Errorf("reading current settings: %w", err)
	}

	backupID := ""
	if s.backup != nil {
		backup, err := s.backup.CreateBackup(ctx, update.Editor, "pre-update")
		if err == nil {
			backupID = backup.ID
		}
	}

	var merged map[string]any
	switch update.Strategy {
	case domain.MergeReplace:
		merged = update.Settings
	default:
		merged = DeepMerge(current.Data, update.Settings)
	}

	diff := ComputeDiff(current.Data, merged)

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("marshaling settings: %w", err)
	}

	if err := editor.WriteSettings(ctx, update.Scope, data); err != nil {
		return nil, "", fmt.Errorf("writing settings: %w", err)
	}

	return diff, backupID, nil
}

func (s *SettingsService) GetEditor(id domain.EditorID) (ports.Editor, bool) {
	e, ok := s.editors[id]
	return e, ok
}

func DeepMerge(base, overlay map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		if baseMap, ok := result[k].(map[string]any); ok {
			if overlayMap, ok := v.(map[string]any); ok {
				result[k] = DeepMerge(baseMap, overlayMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}

func ComputeDiff(before, after map[string]any) *domain.SettingsDiff {
	diff := &domain.SettingsDiff{
		Added:   make(map[string]any),
		Changed: make(map[string]any),
	}

	for k, v := range after {
		old, exists := before[k]
		if !exists {
			diff.Added[k] = v
		} else if !jsonEqual(old, v) {
			diff.Changed[k] = v
		}
	}

	for k := range before {
		if _, exists := after[k]; !exists {
			diff.Removed = append(diff.Removed, k)
		}
	}

	return diff
}

func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
