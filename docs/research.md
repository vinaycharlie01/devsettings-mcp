# DevSettingsMCP — Research Summary

Date: 2026-11-10

## 1. Existing Ecosystem Analysis

### 1.1 Existing MCP Servers for Editor Configuration

| Project | Language | Stars | Status | Capabilities | Limitations |
|---------|----------|-------|--------|-------------|-------------|
| [vscode-mcp-server](https://github.com/markjspivey-xwisee/vscode-mcp-server) | Unknown | 2 | Inactive (last update Feb 2025) | VS Code settings, extensions, snippets management | Single editor only, unmaintained, no migration |
| [Context7](https://github.com/upstash/context7) | Python | N/A | Active (Upstash maintained) | Documentation retrieval for libraries/frameworks | Documentation-only, no settings management |

**Key finding:** No existing MCP server provides cross-editor settings management, migration, or backup/restore. DevSettingsMCP fills a genuine gap.

### 1.2 Related Tools (Non-MCP)

| Tool | Language | Stars | Purpose | Relevance |
|------|----------|-------|---------|-----------|
| [chezmoi](https://github.com/twpayne/chezmoi) | Go | ~13k | Dotfiles management across machines | Manages config files but not IDE-specific settings; no extension management; no migration between editors |
| Nix / Home Manager | Nix | N/A | Declarative system configuration | Powerful but heavy; requires Nix ecosystem adoption; not IDE-focused |
| Dev Containers | JSON | N/A | Containerized dev environments | Development environment specification, not IDE settings migration |
| JetBrains Migrate from Windsurf | Built-in | N/A | One-way migration into JetBrains | IDE-specific, not general-purpose |

### 1.3 Differentiation

DevSettingsMCP is unique in combining:
1. MCP protocol integration (AI agent-accessible)
2. Cross-editor settings management
3. Cross-editor migration with compatibility reporting
4. Backup/restore with secret redaction
5. Documentation retrieval for verified configuration

No existing tool provides all five capabilities.

---

## 2. Official Editor Documentation Research

### 2.1 VS Code

**Source:** https://code.visualstudio.com/docs

#### Settings Files

| Type | macOS | Linux | Windows |
|------|-------|-------|---------|
| User settings | `~/Library/Application Support/Code/User/settings.json` | `~/.config/Code/User/settings.json` | `%APPDATA%\Code\User\settings.json` |
| Workspace settings | `<workspace>/.vscode/settings.json` | Same | Same |
| Keybindings | `~/Library/Application Support/Code/User/keybindings.json` | `~/.config/Code/User/keybindings.json` | `%APPDATA%\Code\User\keybindings.json` |
| Snippets | `~/Library/Application Support/Code/User/snippets/` | `~/.config/Code/User/snippets/` | `%APPDATA%\Code\User\snippets\` |

#### Precedence
Defaults → User settings → Workspace settings → Folder settings (multi-root)

#### Extension Management (CLI)
- `code --list-extensions` — list installed extensions
- `code --list-extensions --show-versions` — with versions
- `code --install-extension <publisher.extension>` — install
- `code --install-extension <path-to-vsix>` — install from file
- `code --uninstall-extension <publisher.extension>` — uninstall
- `code --force --install-extension <id>` — force update

#### Configuration Format
- JSON (with optional comments via JSONC)
- Language-specific overrides via `"[language]": { ... }` blocks
- Settings schema available via VS Code's built-in schema validation

#### Key Settings Areas
- `editor.formatOnSave` — format on save
- `editor.tabSize` — tab size
- `editor.defaultFormatter` — default formatter
- `editor.codeActionsOnSave` — code actions on save
- Language server settings via extension-specific keys (e.g., `gopls`)
- Debug configurations in `.vscode/launch.json`
- Task definitions in `.vscode/tasks.json`

#### MCP Configuration
- VS Code supports MCP server configuration via settings or `.vscode/mcp.json`
- Configurable at user and workspace level

### 2.2 Zed

**Source:** https://zed.dev/docs/configuring-zed

#### Settings Files

| Type | macOS | Linux | Windows |
|------|-------|-------|---------|
| User settings | `~/Library/Application Support/Zed/settings.json` | `~/.config/zed/settings.json` | `%APPDATA%\Zed\settings.json` |
| Project settings | `<project>/.zed/settings.json` | Same | Same |

#### Configuration Format
- JSON with `//` comments supported
- Flat structure with nested objects for specific areas
- Same settings file across all release channels (Stable, Preview, Nightly)

#### Extension Management
- **No CLI for extensions.** Extensions are managed through the Extension Gallery (Cmd+Shift+X / Ctrl+Shift+X).
- `auto_install_extensions` setting in settings.json allows declarative extension management:
  ```json
  { "auto_install_extensions": { "extension-id": true } }
  ```
- Extension install location:
  - macOS: `~/Library/Application Support/Zed/extensions/`
  - Linux: `~/.local/share/zed/extensions/`
  - Windows: `%LOCALAPPDATA%\Zed\extensions\`

#### Key Settings Areas
- `format_on_save` — "on" or "off" (string, not boolean)
- `tab_size` — integer
- `formatter` — formatter configuration
- `lsp` — language server configuration
- `languages` — per-language overrides
- `theme` — active theme
- `vim_mode` — vim keybindings

#### MCP Configuration
- Zed supports MCP server configuration in settings.json under `"context_servers"` key

### 2.3 Cursor

**Source:** https://docs.cursor.com/

#### Settings Files

| Type | macOS | Linux | Windows |
|------|-------|-------|---------|
| IDE settings | `~/Library/Application Support/Cursor/User/settings.json` | `~/.config/Cursor/User/settings.json` | `%APPDATA%\Cursor\User\settings.json` |
| CLI config | `~/.cursor/cli-config.json` | `~/.cursor/cli-config.json` or `$XDG_CONFIG_HOME/cursor/cli-config.json` | `%USERPROFILE%\.cursor\cli-config.json` |
| Project CLI | `<project>/.cursor/cli.json` | Same | Same |

#### Configuration Format
- VS Code-compatible JSON for IDE settings
- Separate CLI configuration files for Cursor Agent CLI

#### Extension Management
- Cursor uses the VS Code extension ecosystem
- Same CLI commands as VS Code: `cursor --list-extensions`, `cursor --install-extension`

#### Key Differences from VS Code
- `.cursor/` directory for project-level agent configuration
- `permissions.json` for tool permissions
- `sandbox.json` for sandbox policy
- Cursor-specific settings in "Cursor Settings" UI panel

### 2.4 Windsurf (now Devin Desktop)

**Source:** https://docs.windsurf.com/

#### Settings Files

| Type | macOS | Linux | Windows |
|------|-------|-------|---------|
| User settings | `~/Library/Application Support/Windsurf/User/settings.json` | `~/.config/Windsurf/User/settings.json` | `%APPDATA%\Windsurf\User\settings.json` |
| MCP config | `~/.codeium/windsurf/mcp_config.json` | Same | Same |
| AI rules | `<project>/.windsurfrules` or `<project>/.windsurf/rules/*.md` | Same | Same |

**Note:** Windsurf was rebranded to Devin Desktop. Settings may migrate to `.devin/` paths. This editor is deprioritized for the MVP.

### 2.5 JetBrains IDEs

**Source:** https://www.jetbrains.com/help/

#### Settings Directories

| Type | macOS | Linux | Windows |
|------|-------|-------|---------|
| Config | `~/Library/Application Support/JetBrains/<product><version>` | `~/.config/JetBrains/<product><version>` | `%APPDATA%\JetBrains\<product><version>` |

- Product names: `IntelliJIdea`, `GoLand`, `WebStorm`, `PyCharm`, `CLion`, etc.
- Version appended: e.g., `IntelliJIdea2026.2`
- Override via `idea.config.path` property

#### Configuration Format
- XML files in subdirectories (options/, keymaps/, colors/)
- Not a single JSON file — fundamentally different from VS Code/Zed/Cursor

#### Extension Management
- **No CLI for plugin installation.** Managed through IDE UI or JetBrains Marketplace.
- Export/import via ZIP archive (File → Manage IDE Settings → Export/Import)
- Backup and Sync plugin for cloud synchronization

**Note:** JetBrains uses XML-based settings in versioned directories with no standardized CLI. This makes it significantly harder to manage programmatically compared to JSON-based editors. JetBrains support is deprioritized for the MVP.

---

## 3. Documentation Sources

### 3.1 Available Documentation Sources for Research Feature

| Source | Format | Access Method | License Consideration |
|--------|--------|---------------|----------------------|
| VS Code docs | Markdown | GitHub: microsoft/vscode-docs | MIT License |
| Zed docs | Markdown | GitHub: zed-industries/zed (docs/) | AGPL-3.0 (docs likely permissive) |
| Go docs | HTML/Markdown | pkg.go.dev, Go wiki | BSD-3-Clause |
| Rust docs | Markdown/HTML | doc.rust-lang.org | MIT/Apache-2.0 |
| Python docs | reStructuredText | docs.python.org | PSF License |
| MCP spec | Markdown | GitHub: modelcontextprotocol/modelcontextprotocol | Apache-2.0 |

### 3.2 Strategy

For the MVP documentation feature:
1. Use pre-configured documentation source URLs
2. Fetch documentation via HTTP where licensing permits
3. Parse Markdown content for search
4. Cache locally on filesystem
5. Return content with source URLs
6. Start with Go and VS Code/Zed documentation
7. Defer full-text indexing (SQLite FTS5) until search volume warrants it

---

## 4. Technology Decisions

### 4.1 Go MCP SDK

**Selected:** `github.com/modelcontextprotocol/go-sdk` v1.8.0

- Official SDK maintained by MCP organization in collaboration with Google
- 5,300+ stars, actively maintained
- Supports MCP protocol 2026-07-28
- stdio transport via `mcp.StdioTransport{}`
- Typed tool registration via `mcp.AddTool()`
- Context cancellation support
- Go release policy compliance

**API Summary:**
```go
server := mcp.NewServer(&mcp.Implementation{Name: "devsettings-mcp", Version: "0.1.0"}, nil)
mcp.AddTool(server, &mcp.Tool{Name: "get_settings", Description: "..."}, handler)
server.Run(ctx, &mcp.StdioTransport{})
```

### 4.2 Mock Generation

**Selected:** `go.uber.org/mock` with `mockgen`

- Actively maintained successor to `github.com/golang/mock`
- Expectation-based mocking with type-safe helpers
- Standard `//go:generate` integration
- Widely used in the Go ecosystem

For simple cases, handwritten fakes will be preferred when clearer.

### 4.3 Task Runner

**Selected:** [Task](https://taskfile.dev/) (Taskfile.yaml)

- Go-based, single binary
- YAML syntax
- Cross-platform (macOS, Linux, Windows)
- CI integration via `go-task/setup-task@v2` GitHub Action

---

## 5. Editor Capability Comparison

### 5.1 Settings Management

| Capability | VS Code | Zed | Cursor |
|-----------|---------|-----|--------|
| Settings format | JSON/JSONC | JSON with comments | JSON/JSONC (VS Code compatible) |
| User settings | Yes | Yes | Yes |
| Workspace/project settings | Yes (.vscode/) | Yes (.zed/) | Yes (.vscode/ + .cursor/) |
| Settings schema | Yes (built-in) | Partial | Yes (VS Code based) |
| Language-specific settings | Yes (`[language]` blocks) | Yes (`languages` key) | Yes (VS Code compatible) |
| Settings precedence | User → Workspace → Folder | User → Project | User → Workspace |

### 5.2 Extension Management

| Capability | VS Code | Zed | Cursor |
|-----------|---------|-----|--------|
| CLI list extensions | `code --list-extensions` | No CLI | `cursor --list-extensions` |
| CLI install extensions | `code --install-extension` | No CLI | `cursor --install-extension` |
| Declarative extensions | No built-in | `auto_install_extensions` | No built-in |
| Extension ecosystem | VS Code Marketplace | Zed Extension Gallery | VS Code Marketplace |
| Extension ID format | `publisher.name` | `name` (no publisher prefix) | `publisher.name` |

### 5.3 Format On Save

| Editor | Setting Key | Values |
|--------|------------|--------|
| VS Code | `editor.formatOnSave` | `true` / `false` |
| Zed | `format_on_save` | `"on"` / `"off"` |
| Cursor | `editor.formatOnSave` | `true` / `false` (VS Code compatible) |

### 5.4 Tab Size

| Editor | Setting Key | Values |
|--------|------------|--------|
| VS Code | `editor.tabSize` | integer |
| Zed | `tab_size` | integer |
| Cursor | `editor.tabSize` | integer (VS Code compatible) |

### 5.5 MCP Server Configuration

| Editor | Location | Format |
|--------|----------|--------|
| VS Code | `settings.json` or `.vscode/mcp.json` | JSON object |
| Zed | `settings.json` under `context_servers` | JSON object |
| Cursor | `.cursor/mcp.json` or settings | JSON object |

---

## 6. Migration Feasibility Analysis

### 6.1 VS Code ↔ Cursor

**Feasibility: High**

Cursor is VS Code-based. Settings are largely compatible:
- Same JSON format
- Same settings keys for most editor features
- Same extension ecosystem
- Same CLI interface

Migration is mostly file copy with Cursor-specific additions.

### 6.2 VS Code → Zed

**Feasibility: Medium**

Key mappings required:
- `editor.formatOnSave: true` → `format_on_save: "on"`
- `editor.tabSize: N` → `tab_size: N`
- `editor.fontSize: N` → `buffer_font_size: N`
- `editor.fontFamily: "..."` → `buffer_font_family: "..."`
- Language-specific settings require restructuring

Challenges:
- Different extension ecosystems (no direct ID mapping)
- Different settings key names and structures
- Some VS Code features have no Zed equivalent
- Zed has no CLI for extension installation (use `auto_install_extensions`)

### 6.3 Zed → VS Code

**Feasibility: Medium**

Reverse of above. Additional challenge: Zed's simpler settings model means some VS Code features won't have source values.

### 6.4 Extension Equivalence (Initial Mapping)

| Purpose | VS Code Extension | Zed Equivalent |
|---------|-------------------|----------------|
| Go | `golang.go` | Built-in Go support |
| Rust | `rust-lang.rust-analyzer` | `rust-analyzer` extension |
| Python | `ms-python.python` | `python` extension |
| Docker | `ms-azuretools.vscode-docker` | `dockerfile` extension |
| YAML | `redhat.vscode-yaml` | Built-in YAML support |
| TOML | `tamasfe.even-better-toml` | `toml` extension |
| GitLens | `eamodio.gitlens` | No equivalent |
| Vim mode | `vscodevim.vim` | Built-in `vim_mode` setting |

**Note:** Extension mappings must be maintained as a verified, curated list. Never auto-generate extension equivalences.

---

## 7. Canonical Configuration Model

### 7.1 Design

The canonical model represents developer intent, not editor-specific JSON.

```yaml
editor:
  format_on_save: true        # boolean
  tab_size: 4                  # integer
  font_size: 14                # integer
  font_family: "JetBrains Mono"  # string
  word_wrap: true              # boolean
  minimap: false               # boolean
  line_numbers: "relative"     # "on" | "off" | "relative"
  vim_mode: false              # boolean

languages:
  go:
    formatter: "gofumpt"
    lsp: "gopls"
    tab_size: 4
    format_on_save: true
  python:
    formatter: "black"
    lsp: "pyright"
    tab_size: 4
  rust:
    formatter: "rustfmt"
    lsp: "rust-analyzer"
    tab_size: 4

theme: "One Dark Pro"

extensions:
  - id: "docker"
    purpose: "Docker support"
  - id: "yaml"
    purpose: "YAML support"
```

### 7.2 Mapping Rules

Each adapter maps canonical fields to native settings:

| Canonical | VS Code | Zed |
|-----------|---------|-----|
| `editor.format_on_save` | `editor.formatOnSave: true` | `format_on_save: "on"` |
| `editor.tab_size` | `editor.tabSize: 4` | `tab_size: 4` |
| `editor.font_size` | `editor.fontSize: 14` | `buffer_font_size: 14` |
| `editor.vim_mode` | Extension `vscodevim.vim` | `vim_mode: true` |
| `languages.go.lsp` | Extension `golang.go` + gopls config | Built-in + `lsp.gopls` config |

---

## 8. Security Considerations

1. **Read-only by default** — All read operations require no approval
2. **Plan before write** — Generate a diff/plan before modifying settings
3. **Approval required** — User must approve before any file write
4. **Backup before modify** — Automatic backup before any settings change
5. **Path validation** — Prevent path traversal; validate all paths are within expected config directories
6. **Secret redaction** — Scan for and redact API keys, tokens, passwords in backups/exports
7. **Atomic writes** — Write to temp file, then rename (prevents partial writes)
8. **No shell execution** — Never execute arbitrary commands from settings or documentation content
9. **Documentation is untrusted** — Retrieved docs are data, never executed
10. **Extension metadata is untrusted** — Never auto-execute instructions from extension descriptions
11. **Rollback support** — Every modification can be reverted from backup

### Secret Detection Patterns
```
- API keys: /(?i)(api[_-]?key|apikey)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}/
- Tokens: /(?i)(token|bearer|auth)\s*[:=]\s*["']?[A-Za-z0-9_\-\.]{20,}/
- Passwords: /(?i)(password|passwd|pwd)\s*[:=]\s*["']?.+/
- Private keys: /-----BEGIN .* PRIVATE KEY-----/
- AWS keys: /(?i)AKIA[0-9A-Z]{16}/
```

---

## 9. MCP Tool Definitions

### 9.1 Settings Tools

#### detect_editors
Detect installed editors and their configuration paths.

**Input:** None (or optional OS override for testing)
**Output:**
```json
{
  "editors": [
    {
      "name": "vscode",
      "display_name": "Visual Studio Code",
      "installed": true,
      "version": "1.95.0",
      "config_dir": "/home/user/.config/Code/User",
      "settings_path": "/home/user/.config/Code/User/settings.json"
    }
  ]
}
```

#### get_settings
Read current settings for an editor.

**Input:**
```json
{
  "editor": "vscode",
  "scope": "user"  // "user" | "workspace"
}
```

**Output:** Current settings JSON with metadata.

#### update_settings
Update settings for an editor (requires approval).

**Input:**
```json
{
  "editor": "vscode",
  "scope": "user",
  "settings": { "editor.formatOnSave": true },
  "merge_strategy": "merge"  // "merge" | "replace"
}
```

**Output:** Diff of changes, backup ID, success status.

### 9.2 Documentation Tools

#### search_docs
Search documentation for a topic.

**Input:**
```json
{
  "query": "format on save configuration",
  "source": "vscode"  // optional filter
}
```

**Output:** List of matching documentation sections with URLs.

#### fetch_docs
Retrieve a specific documentation page.

**Input:**
```json
{
  "url": "https://code.visualstudio.com/docs/getstarted/settings",
  "section": "format on save"  // optional section filter
}
```

**Output:** Documentation content with source reference.

### 9.3 Backup Tools

#### create_backup
Create a backup of editor configuration.

**Input:**
```json
{
  "editor": "vscode",
  "label": "before-go-setup"  // optional
}
```

**Output:** Backup manifest with ID, timestamp, file list.

#### restore_backup
Restore a previous backup.

**Input:**
```json
{
  "backup_id": "vscode-2026-11-10-001",
  "dry_run": true  // preview changes before applying
}
```

**Output:** Diff of changes to apply, or success confirmation.

### 9.4 Migration Tools

#### plan_migration
Generate a migration plan between editors.

**Input:**
```json
{
  "source": "vscode",
  "target": "zed"
}
```

**Output:** Migration plan with compatibility report.

#### execute_migration
Execute an approved migration plan.

**Input:**
```json
{
  "plan_id": "migration-vscode-zed-2026-11-10",
  "approved": true
}
```

**Output:** Migration report with results per item.

---

## 10. Architecture Diagram

```
┌─────────────────────────────────────────────┐
│              AI Coding Agent                 │
│         (Claude, Copilot, etc.)              │
└────────────────┬────────────────────────────┘
                 │ MCP Protocol (stdio)
                 ▼
┌─────────────────────────────────────────────┐
│           MCP Server Layer                   │
│  internal/mcp/server.go, tools.go            │
│  Tool handlers (thin, delegate to services)  │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│         Application Services                 │
│  internal/application/                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │ Settings  │ │ Backup   │ │Migration │    │
│  │ Service   │ │ Service  │ │ Service  │    │
│  └──────────┘ └──────────┘ └──────────┘    │
│  ┌──────────────┐                           │
│  │Documentation │                           │
│  │  Service     │                           │
│  └──────────────┘                           │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│           Domain / Ports                     │
│  internal/domain/ + internal/ports/          │
│  Editor interface, Settings, Migration types │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│         Editor Adapters                      │
│  internal/adapters/                          │
│  ┌────────┐  ┌──────┐  ┌────────┐          │
│  │ VS Code│  │ Zed  │  │ Cursor │          │
│  │Adapter │  │Adapt.│  │Adapter │          │
│  └────────┘  └──────┘  └────────┘          │
│         ▼         ▼         ▼               │
│    [Filesystem: settings.json, extensions]   │
└─────────────────────────────────────────────┘
```

---

## 11. MVP Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [x] Research complete
- [ ] Go module initialization
- [ ] MCP server with stdio transport
- [ ] detect_editors tool (VS Code, Zed, Cursor detection)
- [ ] get_settings tool (read settings files)
- [ ] Taskfile.yaml with build, test, lint, fmt tasks
- [ ] Unit tests for editor detection and settings reading
- [ ] CI pipeline (task ci)

### Phase 2: Settings Management (Week 2)
- [ ] Settings merge logic (deep merge with conflict detection)
- [ ] update_settings tool with plan/approve/apply workflow
- [ ] Settings validation
- [ ] create_backup tool (automatic before writes)
- [ ] restore_backup tool
- [ ] Secret redaction in backups
- [ ] Mock generation for editor interface
- [ ] Unit tests for merge, validation, backup, redaction

### Phase 3: Documentation (Week 3)
- [ ] Documentation source registry
- [ ] search_docs tool (keyword search over indexed docs)
- [ ] fetch_docs tool (retrieve specific pages)
- [ ] Local filesystem cache
- [ ] Source URL references in all responses
- [ ] Documentation tests

### Phase 4: Editor Adapters (Week 3-4)
- [ ] VS Code adapter (full implementation)
- [ ] Zed adapter (full implementation)
- [ ] Cursor adapter (extends VS Code adapter)
- [ ] Per-adapter unit tests with filesystem fixtures

### Phase 5: Migration (Week 4-5)
- [ ] Canonical configuration model
- [ ] VS Code → Zed migration with compatibility report
- [ ] Zed → VS Code migration
- [ ] VS Code ↔ Cursor migration
- [ ] plan_migration tool
- [ ] execute_migration tool
- [ ] Migration rollback
- [ ] Integration tests

### Phase 6: Release (Week 5-6)
- [ ] Cross-platform build (macOS, Linux, Windows)
- [ ] Example profiles (go-backend, rust, python, web-frontend)
- [ ] README with installation and usage instructions
- [ ] GitHub Actions CI/CD
- [ ] Integration tests in CI
- [ ] Open-source release preparation

---

## 12. Assumptions and Verification Notes

### Verified Facts
- VS Code settings paths confirmed via official docs
- VS Code CLI extension management confirmed via official docs
- Zed settings paths confirmed via official docs
- Zed `auto_install_extensions` confirmed via official docs
- Cursor settings paths confirmed via docs and community sources
- Go MCP SDK v1.8.0 exists with stdio transport support
- chezmoi is MIT licensed, written in Go, single binary

### Assumptions (Need Verification)
- Zed extension install directory paths (from third-party sources)
- Cursor CLI commands mirror VS Code CLI (reasonable assumption, needs testing)
- Windsurf/Devin Desktop paths may change with rebrand
- JetBrains config directory format for 2026.2 versions
- Some Zed settings keys inferred from search results (need schema verification)

### Explicitly Not Supported in MVP
- JetBrains IDEs (XML-based config, no CLI, complex versioned directories)
- Windsurf/Devin Desktop (rebrand in progress, unstable paths)
- Neovim (Lua-based config, different paradigm)
- Sublime Text (deprioritized)
- Extension auto-discovery/recommendation
- Cloud sync
- Vector database for documentation
- Remote/SSH editor instances
