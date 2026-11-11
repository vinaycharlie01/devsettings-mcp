package domain

type CanonicalAI struct {
	InlineCompletion  bool   `json:"inline_completion" yaml:"inline_completion"`
	ChatEnabled       bool   `json:"chat_enabled" yaml:"chat_enabled"`
	Model             string `json:"model,omitempty" yaml:"model,omitempty"`
	Provider          string `json:"provider,omitempty" yaml:"provider,omitempty"`
	ComposerEnabled   bool   `json:"composer_enabled,omitempty" yaml:"composer_enabled,omitempty"`
	PredictionEnabled bool   `json:"prediction_enabled,omitempty" yaml:"prediction_enabled,omitempty"`
}

type MCPServer struct {
	Name      string            `json:"name"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	URL       string            `json:"url,omitempty"`
	Transport string            `json:"transport,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
}

type MCPConfig struct {
	Editor  EditorID    `json:"editor"`
	Servers []MCPServer `json:"servers"`
	Path    string      `json:"path"`
}

// AISettingsMapping shows how AI settings map between editors.
type AISettingsMapping struct {
	CanonicalField string `json:"canonical_field"`
	VSCodeKey      string `json:"vscode_key,omitempty"`
	ZedKey         string `json:"zed_key,omitempty"`
	CursorKey      string `json:"cursor_key,omitempty"`
	Note           string `json:"note,omitempty"`
}

func AISettingsMappings() []AISettingsMapping {
	return []AISettingsMapping{
		{
			CanonicalField: "inline_completion",
			VSCodeKey:      "github.copilot.enable",
			ZedKey:         "features.inline_completion_provider",
			CursorKey:      "cursor.tab.enabled",
		},
		{
			CanonicalField: "chat_enabled",
			VSCodeKey:      "github.copilot.chat.enabled",
			ZedKey:         "assistant.enabled",
			CursorKey:      "cursor.chat.enabled",
		},
		{
			CanonicalField: "model",
			VSCodeKey:      "github.copilot.advanced.model",
			ZedKey:         "assistant.default_model.model",
			CursorKey:      "cursor.ai.model",
		},
		{
			CanonicalField: "provider",
			VSCodeKey:      "",
			ZedKey:         "assistant.default_model.provider",
			CursorKey:      "",
			Note:           "Provider is Zed-specific; VS Code/Cursor infer from model",
		},
		{
			CanonicalField: "composer_enabled",
			VSCodeKey:      "",
			ZedKey:         "",
			CursorKey:      "cursor.composer.enabled",
			Note:           "Composer is Cursor-specific",
		},
		{
			CanonicalField: "prediction_enabled",
			VSCodeKey:      "",
			ZedKey:         "",
			CursorKey:      "cursor.prediction.enabled",
			Note:           "Prediction is Cursor-specific",
		},
	}
}
