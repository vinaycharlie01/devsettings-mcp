package domain

type Profile struct {
	Name        string                     `yaml:"name" json:"name"`
	Description string                     `yaml:"description" json:"description"`
	Languages   map[string]ProfileLanguage `yaml:"languages,omitempty" json:"languages,omitempty"`
	Extensions  []string                   `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Settings    map[string]any             `yaml:"settings,omitempty" json:"settings,omitempty"`
}

type ProfileLanguage struct {
	LSP       string `yaml:"lsp,omitempty" json:"lsp,omitempty"`
	Formatter string `yaml:"formatter,omitempty" json:"formatter,omitempty"`
	Debugger  string `yaml:"debugger,omitempty" json:"debugger,omitempty"`
}
