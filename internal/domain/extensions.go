package domain

type ExtensionMapping struct {
	Purpose  string `json:"purpose"`
	Category string `json:"category"`
	VSCode   string `json:"vscode"`
	Zed      string `json:"zed,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

func (m *ExtensionMapping) ExtensionID(editor EditorID) string {
	switch editor {
	case VSCode:
		return m.VSCode
	case Zed:
		return m.Zed
	case Cursor:
		return m.Cursor
	default:
		return ""
	}
}

func FindEquivalent(sourceEditor, targetEditor EditorID, extensionID string) (string, *ExtensionMapping) {
	for i, m := range ExtensionEquivalences {
		if m.ExtensionID(sourceEditor) == extensionID {
			targetID := m.ExtensionID(targetEditor)
			return targetID, &ExtensionEquivalences[i]
		}
	}
	return "", nil
}

var ExtensionEquivalences = []ExtensionMapping{
	{Purpose: "Go support", Category: "language", VSCode: "golang.go", Zed: "go", Cursor: "golang.go"},
	{Purpose: "Rust support", Category: "language", VSCode: "rust-lang.rust-analyzer", Zed: "rust", Cursor: "rust-lang.rust-analyzer"},
	{Purpose: "Python support", Category: "language", VSCode: "ms-python.python", Zed: "python", Cursor: "ms-python.python"},
	{Purpose: "TypeScript/JavaScript", Category: "language", VSCode: "ms-vscode.vscode-typescript-next", Zed: "javascript", Cursor: "ms-vscode.vscode-typescript-next", Notes: "Zed has built-in TS support"},
	{Purpose: "C/C++ support", Category: "language", VSCode: "ms-vscode.cpptools", Zed: "c", Cursor: "ms-vscode.cpptools"},
	{Purpose: "Java support", Category: "language", VSCode: "redhat.java", Zed: "java", Cursor: "redhat.java"},
	{Purpose: "Prettier formatter", Category: "formatter", VSCode: "esbenp.prettier-vscode", Zed: "prettier", Cursor: "esbenp.prettier-vscode"},
	{Purpose: "ESLint", Category: "linter", VSCode: "dbaeumer.vscode-eslint", Zed: "eslint", Cursor: "dbaeumer.vscode-eslint"},
	{Purpose: "Git Lens", Category: "git", VSCode: "eamodio.gitlens", Cursor: "eamodio.gitlens", Notes: "No Zed equivalent"},
	{Purpose: "One Dark Pro theme", Category: "theme", VSCode: "zhuangtongfa.Material-theme", Zed: "one-dark-pro", Cursor: "zhuangtongfa.Material-theme"},
	{Purpose: "Catppuccin theme", Category: "theme", VSCode: "Catppuccin.catppuccin-vsc", Zed: "catppuccin", Cursor: "Catppuccin.catppuccin-vsc"},
	{Purpose: "TOML support", Category: "file-type", VSCode: "tamasfe.even-better-toml", Zed: "toml", Cursor: "tamasfe.even-better-toml"},
	{Purpose: "YAML support", Category: "file-type", VSCode: "redhat.vscode-yaml", Zed: "yaml", Cursor: "redhat.vscode-yaml"},
	{Purpose: "Docker support", Category: "file-type", VSCode: "ms-azuretools.vscode-docker", Zed: "dockerfile", Cursor: "ms-azuretools.vscode-docker"},
	{Purpose: "Vim emulation", Category: "editor", VSCode: "vscodevim.vim", Cursor: "vscodevim.vim", Notes: "Zed has built-in vim_mode"},
	{Purpose: "Error Lens", Category: "diagnostics", VSCode: "usernamehw.errorlens", Cursor: "usernamehw.errorlens", Notes: "Zed shows inline diagnostics natively"},
	{Purpose: "Markdown support", Category: "file-type", VSCode: "yzhang.markdown-all-in-one", Zed: "markdown", Cursor: "yzhang.markdown-all-in-one"},
	{Purpose: "Test Explorer", Category: "testing", VSCode: "hbenl.vscode-test-explorer", Cursor: "hbenl.vscode-test-explorer", Notes: "No Zed equivalent"},
	{Purpose: "GraphQL support", Category: "language", VSCode: "GraphQL.vscode-graphql", Zed: "graphql", Cursor: "GraphQL.vscode-graphql"},
	{Purpose: "Tailwind CSS", Category: "framework", VSCode: "bradlc.vscode-tailwindcss", Zed: "tailwindcss", Cursor: "bradlc.vscode-tailwindcss"},
}
