package domain

type CompatibilityStatus string

const (
	FullySupported       CompatibilityStatus = "fully_supported"
	SupportedTransformed CompatibilityStatus = "supported_with_transformation"
	EquivalentExtension  CompatibilityStatus = "equivalent_extension_available"
	PartiallySupported   CompatibilityStatus = "partially_supported"
	Unsupported          CompatibilityStatus = "unsupported"
	ManualIntervention   CompatibilityStatus = "requires_manual_intervention"
	ExcludedSecurity     CompatibilityStatus = "excluded_for_security"
)

type MigrationItem struct {
	SourceKey   string              `json:"source_key"`
	TargetKey   string              `json:"target_key,omitempty"`
	SourceValue any                 `json:"source_value"`
	TargetValue any                 `json:"target_value,omitempty"`
	Status      CompatibilityStatus `json:"status"`
	Note        string              `json:"note,omitempty"`
}

type MigrationPlan struct {
	ID       string          `json:"id"`
	Source   EditorID        `json:"source"`
	Target   EditorID        `json:"target"`
	Items    []MigrationItem `json:"items"`
	BackupID string          `json:"backup_id"`
}

type CanonicalSettings struct {
	Editor     CanonicalEditor              `yaml:"editor" json:"editor"`
	Languages  map[string]CanonicalLanguage `yaml:"languages,omitempty" json:"languages,omitempty"`
	Theme      string                       `yaml:"theme,omitempty" json:"theme,omitempty"`
	Extensions []CanonicalExtension         `yaml:"extensions,omitempty" json:"extensions,omitempty"`
}

type CanonicalEditor struct {
	FormatOnSave bool   `yaml:"format_on_save" json:"format_on_save"`
	TabSize      int    `yaml:"tab_size" json:"tab_size"`
	FontSize     int    `yaml:"font_size,omitempty" json:"font_size,omitempty"`
	FontFamily   string `yaml:"font_family,omitempty" json:"font_family,omitempty"`
	WordWrap     bool   `yaml:"word_wrap,omitempty" json:"word_wrap,omitempty"`
	LineNumbers  string `yaml:"line_numbers,omitempty" json:"line_numbers,omitempty"`
	VimMode      bool   `yaml:"vim_mode,omitempty" json:"vim_mode,omitempty"`
}

type CanonicalLanguage struct {
	Formatter    string `yaml:"formatter,omitempty" json:"formatter,omitempty"`
	LSP          string `yaml:"lsp,omitempty" json:"lsp,omitempty"`
	Debugger     string `yaml:"debugger,omitempty" json:"debugger,omitempty"`
	TabSize      int    `yaml:"tab_size,omitempty" json:"tab_size,omitempty"`
	FormatOnSave *bool  `yaml:"format_on_save,omitempty" json:"format_on_save,omitempty"`
}

type CanonicalExtension struct {
	ID      string `yaml:"id" json:"id"`
	Purpose string `yaml:"purpose,omitempty" json:"purpose,omitempty"`
}
