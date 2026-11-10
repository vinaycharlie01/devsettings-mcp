package domain

import "runtime"

type EditorID string

const (
	VSCode EditorID = "vscode"
	Zed    EditorID = "zed"
	Cursor EditorID = "cursor"
)

type EditorInfo struct {
	ID           EditorID `json:"id"`
	DisplayName  string   `json:"display_name"`
	Installed    bool     `json:"installed"`
	ConfigDir    string   `json:"config_dir"`
	SettingsPath string   `json:"settings_path"`
}

type Extension struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func CurrentOS() string {
	return runtime.GOOS
}
