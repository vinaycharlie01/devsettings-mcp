package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type DocsService struct {
	cacheDir string
	index    []domain.DocEntry
}

func NewDocsService(dataDir string) *DocsService {
	svc := &DocsService{
		cacheDir: filepath.Join(dataDir, "docs-cache"),
	}
	svc.index = defaultDocIndex()
	return svc
}

func (s *DocsService) Search(_ context.Context, query string, editorFilter string) (*domain.DocSearchResult, error) {
	query = strings.ToLower(query)
	seen := make(map[string]bool)
	var matches []domain.DocEntry

	for _, entry := range s.index {
		if editorFilter != "" && string(entry.Editor) != editorFilter {
			continue
		}
		if matchesEntry(entry, query) && !seen[entry.URL] {
			seen[entry.URL] = true
			matches = append(matches, entry)
		}
	}

	return &domain.DocSearchResult{Entries: matches, Query: query}, nil
}

func (s *DocsService) Fetch(_ context.Context, url string, section string) (*domain.DocEntry, error) {
	cached, err := s.loadFromCache(url)
	if err == nil {
		if section != "" {
			cached = filterSection(cached, section)
		}
		return cached, nil
	}

	for _, entry := range s.index {
		if entry.URL == url {
			result := entry
			if section != "" {
				return filterSection(&result, section), nil
			}
			return &result, nil
		}
	}

	return nil, fmt.Errorf("documentation page not found: %s (use search_docs to find available pages)", url)
}

func (s *DocsService) CacheDoc(entry *domain.DocEntry) error {
	if err := os.MkdirAll(s.cacheDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	filename := urlToFilename(entry.URL)
	return os.WriteFile(filepath.Join(s.cacheDir, filename), data, 0o644)
}

func (s *DocsService) loadFromCache(url string) (*domain.DocEntry, error) {
	filename := urlToFilename(url)
	data, err := os.ReadFile(filepath.Join(s.cacheDir, filename))
	if err != nil {
		return nil, err
	}
	var entry domain.DocEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func matchesEntry(entry domain.DocEntry, query string) bool {
	if strings.Contains(strings.ToLower(entry.Title), query) ||
		strings.Contains(strings.ToLower(entry.Topic), query) {
		return true
	}
	for _, s := range entry.Sections {
		if strings.Contains(strings.ToLower(s.Heading), query) ||
			strings.Contains(strings.ToLower(s.Content), query) {
			return true
		}
	}
	return false
}

func filterSection(entry *domain.DocEntry, section string) *domain.DocEntry {
	section = strings.ToLower(section)
	filtered := *entry
	filtered.Sections = nil
	for _, s := range entry.Sections {
		if strings.Contains(strings.ToLower(s.Heading), section) {
			filtered.Sections = append(filtered.Sections, s)
		}
	}
	return &filtered
}

func urlToFilename(url string) string {
	r := strings.NewReplacer("/", "_", ":", "_", "?", "_", "#", "_")
	return r.Replace(url) + ".json"
}

func defaultDocIndex() []domain.DocEntry {
	return []domain.DocEntry{
		{
			Title: "VS Code Settings Reference", URL: "https://code.visualstudio.com/docs/getstarted/settings",
			Editor: domain.VSCode, Topic: "settings",
			Sections: []domain.DocSection{
				{Heading: "User and Workspace Settings", Content: "VS Code provides two scopes for settings: User Settings apply globally, Workspace Settings apply to the current workspace."},
				{Heading: "Settings Editor", Content: "Use the Settings editor to review and change VS Code settings. Open with Ctrl+, (Cmd+, on macOS)."},
				{Heading: "Default Settings", Content: "VS Code provides default values for all settings. Modified settings are stored in settings.json."},
			},
		},
		{
			Title: "VS Code Keybindings", URL: "https://code.visualstudio.com/docs/getstarted/keybindings",
			Editor: domain.VSCode, Topic: "keybindings",
			Sections: []domain.DocSection{
				{Heading: "Keyboard Shortcuts Editor", Content: "Open the Keyboard Shortcuts editor with Ctrl+K Ctrl+S to view and modify keybindings."},
			},
		},
		{
			Title: "VS Code Extensions API", URL: "https://code.visualstudio.com/api",
			Editor: domain.VSCode, Topic: "extensions",
			Sections: []domain.DocSection{
				{Heading: "Extension Marketplace", Content: "Browse and install extensions from the VS Code Marketplace."},
			},
		},
		{
			Title: "VS Code Go Language", URL: "https://code.visualstudio.com/docs/languages/go",
			Editor: domain.VSCode, Topic: "go",
			Sections: []domain.DocSection{
				{Heading: "Getting Started", Content: "Install the Go extension for VS Code. It uses gopls for language features."},
				{Heading: "Debugging", Content: "The Go extension supports debugging with Delve (dlv)."},
			},
		},
		{
			Title: "VS Code Rust Language", URL: "https://code.visualstudio.com/docs/languages/rust",
			Editor: domain.VSCode, Topic: "rust",
			Sections: []domain.DocSection{
				{Heading: "rust-analyzer", Content: "Install rust-analyzer extension for Rust language support including completion, diagnostics, and refactoring."},
			},
		},
		{
			Title: "VS Code Python Tutorial", URL: "https://code.visualstudio.com/docs/python/python-tutorial",
			Editor: domain.VSCode, Topic: "python",
			Sections: []domain.DocSection{
				{Heading: "Python Extension", Content: "The Python extension provides rich support including IntelliSense, linting, debugging, and Jupyter notebook support."},
			},
		},
		{
			Title: "VS Code TypeScript Tutorial", URL: "https://code.visualstudio.com/docs/typescript/typescript-tutorial",
			Editor: domain.VSCode, Topic: "typescript",
			Sections: []domain.DocSection{
				{Heading: "Built-in Support", Content: "VS Code has built-in TypeScript support with IntelliSense, refactoring, and navigation."},
			},
		},
		{
			Title: "Zed Configuration", URL: "https://zed.dev/docs/configuring-zed",
			Editor: domain.Zed, Topic: "settings",
			Sections: []domain.DocSection{
				{Heading: "Settings File", Content: "Zed stores settings in ~/.config/zed/settings.json (Linux) or ~/Library/Application Support/Zed/settings.json (macOS)."},
				{Heading: "JSON with Comments", Content: "Zed settings files support // comments in JSON, which are stripped before parsing."},
				{Heading: "Key Bindings", Content: "Customize key bindings in keymap.json alongside settings.json."},
			},
		},
		{
			Title: "Zed Extensions", URL: "https://zed.dev/docs/extensions",
			Editor: domain.Zed, Topic: "extensions",
			Sections: []domain.DocSection{
				{Heading: "Installing Extensions", Content: "Install extensions from the Zed extension registry using the Extensions panel."},
				{Heading: "Language Support", Content: "Zed extensions provide language support including syntax highlighting, LSP integration, and tree-sitter grammars."},
			},
		},
		{
			Title: "Zed Languages", URL: "https://zed.dev/docs/languages",
			Editor: domain.Zed, Topic: "languages",
			Sections: []domain.DocSection{
				{Heading: "Language Server Protocol", Content: "Zed uses LSP for language intelligence. Configure language servers per-language in settings."},
				{Heading: "Supported Languages", Content: "Zed supports Go, Rust, Python, TypeScript, JavaScript, C/C++, and many more via extensions."},
			},
		},
		{
			Title: "Cursor Settings", URL: "https://docs.cursor.com/get-started/settings",
			Editor: domain.Cursor, Topic: "settings",
			Sections: []domain.DocSection{
				{Heading: "VS Code Compatibility", Content: "Cursor is built on VS Code and supports all VS Code settings in settings.json."},
				{Heading: "AI Settings", Content: "Cursor adds AI-specific settings like cursor.tab.enabled, cursor.chat, and model configuration."},
			},
		},
		{
			Title: "Cursor AI Features", URL: "https://docs.cursor.com/features",
			Editor: domain.Cursor, Topic: "ai",
			Sections: []domain.DocSection{
				{Heading: "Cursor Tab", Content: "Cursor Tab provides AI-powered code completion that predicts multi-line edits."},
				{Heading: "Chat", Content: "Cursor Chat lets you ask questions about your codebase with full context awareness."},
				{Heading: "Composer", Content: "Composer enables multi-file edits guided by natural language instructions."},
			},
		},
	}
}
