package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
	"github.com/vinaycharlie01/devsettings-mcp/internal/ports"
)

type MigrationService struct {
	editors     map[domain.EditorID]ports.Editor
	normalizers map[domain.EditorID]ports.SettingsNormalizer
	backup      *BackupService
	plans       map[string]*domain.MigrationPlan
}

func NewMigrationService(
	editors map[domain.EditorID]ports.Editor,
	normalizers map[domain.EditorID]ports.SettingsNormalizer,
	backup *BackupService,
) *MigrationService {
	return &MigrationService{
		editors:     editors,
		normalizers: normalizers,
		backup:      backup,
		plans:       make(map[string]*domain.MigrationPlan),
	}
}

func (s *MigrationService) PlanMigration(ctx context.Context, sourceID, targetID domain.EditorID) (*domain.MigrationPlan, error) {
	sourceEditor, ok := s.editors[sourceID]
	if !ok {
		return nil, fmt.Errorf("unsupported source editor: %s", sourceID)
	}

	if _, ok := s.editors[targetID]; !ok {
		return nil, fmt.Errorf("unsupported target editor: %s", targetID)
	}

	sourceNorm, ok := s.normalizers[sourceID]
	if !ok {
		return nil, fmt.Errorf("no normalizer for source editor: %s", sourceID)
	}

	targetNorm, ok := s.normalizers[targetID]
	if !ok {
		return nil, fmt.Errorf("no normalizer for target editor: %s", targetID)
	}

	sourceSettings, err := sourceEditor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return nil, fmt.Errorf("reading source settings: %w", err)
	}

	canonical, sourceItems, err := sourceNorm.ToCanonical(ctx, sourceSettings)
	if err != nil {
		return nil, fmt.Errorf("normalizing source settings: %w", err)
	}

	_, targetItems, err := targetNorm.FromCanonical(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("converting to target settings: %w", err)
	}

	var allItems []domain.MigrationItem
	allItems = append(allItems, sourceItems...)
	allItems = append(allItems, targetItems...)

	backupID := ""
	if s.backup != nil {
		backup, err := s.backup.CreateBackup(ctx, sourceID, "pre-migration")
		if err == nil {
			backupID = backup.ID
		}
	}

	planID := fmt.Sprintf("migration-%s-%s-%s", sourceID, targetID, time.Now().Format("2006-01-02-150405"))

	plan := &domain.MigrationPlan{
		ID:       planID,
		Source:   sourceID,
		Target:   targetID,
		Items:    allItems,
		BackupID: backupID,
	}

	s.plans[planID] = plan
	return plan, nil
}

func (s *MigrationService) ExecuteMigration(ctx context.Context, planID string) error {
	plan, ok := s.plans[planID]
	if !ok {
		return fmt.Errorf("migration plan not found: %s", planID)
	}

	sourceEditor, ok := s.editors[plan.Source]
	if !ok {
		return fmt.Errorf("source editor not available: %s", plan.Source)
	}

	targetEditor, ok := s.editors[plan.Target]
	if !ok {
		return fmt.Errorf("target editor not available: %s", plan.Target)
	}

	sourceNorm, ok := s.normalizers[plan.Source]
	if !ok {
		return fmt.Errorf("no normalizer for source: %s", plan.Source)
	}

	targetNorm, ok := s.normalizers[plan.Target]
	if !ok {
		return fmt.Errorf("no normalizer for target: %s", plan.Target)
	}

	if s.backup != nil {
		if _, err := s.backup.CreateBackup(ctx, plan.Target, "pre-migration-target"); err != nil {
			return fmt.Errorf("backing up target: %w", err)
		}
	}

	sourceSettings, err := sourceEditor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return fmt.Errorf("reading source settings: %w", err)
	}

	canonical, _, err := sourceNorm.ToCanonical(ctx, sourceSettings)
	if err != nil {
		return fmt.Errorf("normalizing source settings: %w", err)
	}

	targetSettings, _, err := targetNorm.FromCanonical(ctx, canonical)
	if err != nil {
		return fmt.Errorf("converting to target format: %w", err)
	}

	existingTarget, err := targetEditor.ReadSettings(ctx, domain.ScopeUser)
	if err != nil {
		return fmt.Errorf("reading existing target settings: %w", err)
	}

	merged := DeepMerge(existingTarget.Data, targetSettings)

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling target settings: %w", err)
	}

	if err := targetEditor.WriteSettings(ctx, domain.ScopeUser, data); err != nil {
		return fmt.Errorf("writing target settings: %w", err)
	}

	return nil
}
