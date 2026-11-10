package domain

// MCPManifest is the canonical YAML format for MCP server definitions.
// Write once, deploy to all editors — like Helm charts for MCP configs.
type MCPManifest struct {
	Version string              `yaml:"version" json:"version"`
	Servers []MCPManifestServer `yaml:"servers" json:"servers"`
	Vars    map[string]string   `yaml:"vars,omitempty" json:"vars,omitempty"`
}

type MCPManifestServer struct {
	Name      string                         `yaml:"name" json:"name"`
	Command   string                         `yaml:"command,omitempty" json:"command,omitempty"`
	Args      []string                       `yaml:"args,omitempty" json:"args,omitempty"`
	URL       string                         `yaml:"url,omitempty" json:"url,omitempty"`
	Transport string                         `yaml:"transport,omitempty" json:"transport,omitempty"`
	Env       map[string]string              `yaml:"env,omitempty" json:"env,omitempty"`
	Disabled  []EditorID                     `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	Overrides map[EditorID]MCPServerOverride `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

type MCPServerOverride struct {
	Command string            `yaml:"command,omitempty" json:"command,omitempty"`
	Args    []string          `yaml:"args,omitempty" json:"args,omitempty"`
	URL     string            `yaml:"url,omitempty" json:"url,omitempty"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

type ManifestSyncResult struct {
	Editor  EditorID `json:"editor"`
	Path    string   `json:"path"`
	Servers int      `json:"servers"`
	Skipped int      `json:"skipped"`
	Status  string   `json:"status"`
}
