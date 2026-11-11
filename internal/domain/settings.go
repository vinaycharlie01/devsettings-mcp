package domain

import "time"

type SettingsScope string

const (
	ScopeUser      SettingsScope = "user"
	ScopeWorkspace SettingsScope = "workspace"
)

type MergeStrategy string

const (
	MergeDeep    MergeStrategy = "merge"
	MergeReplace MergeStrategy = "replace"
)

type Settings struct {
	Editor EditorID       `json:"editor"`
	Scope  SettingsScope  `json:"scope"`
	Path   string         `json:"path"`
	Data   map[string]any `json:"data"`
}

type SettingsUpdate struct {
	Editor   EditorID       `json:"editor"`
	Scope    SettingsScope  `json:"scope"`
	Settings map[string]any `json:"settings"`
	Strategy MergeStrategy  `json:"merge_strategy"`
}

type SettingsDiff struct {
	Added   map[string]any `json:"added,omitempty"`
	Changed map[string]any `json:"changed,omitempty"`
	Removed []string       `json:"removed,omitempty"`
}

type Backup struct {
	ID             string         `json:"id"`
	Editor         EditorID       `json:"editor"`
	Label          string         `json:"label,omitempty"`
	Timestamp      time.Time      `json:"timestamp"`
	SettingsData   map[string]any `json:"settings_data,omitempty"`
	ExtensionsData []Extension    `json:"extensions_data,omitempty"`
	Checksum       string         `json:"checksum,omitempty"`
}
