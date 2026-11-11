# DevSettingsMCP

> Let your AI agent manage, understand and migrate your development environment.

DevSettingsMCP is an MCP server that allows AI coding agents to manage developer configurations across different IDEs. Built in Go with a clean hexagonal architecture.

## Features

- **Editor Detection** — Automatically detect installed editors (VS Code, Zed, Cursor) and their configuration paths
- **Settings Management** — Read, update, and merge settings with deep merge support
- **Backup & Restore** — Create and restore configuration backups with automatic secret redaction
- **Cross-Editor Migration** — Migrate settings between editors with compatibility reporting
- **Safety First** — Atomic writes, automatic backups before changes, secret redaction, path validation

## Supported Editors

| Editor | Settings | Extensions | Migration |
|--------|----------|------------|-----------|
| VS Code | Read/Write | List/Install/Uninstall | Source & Target |
| Zed | Read/Write | List | Source & Target |
| Cursor | Read/Write | List/Install/Uninstall | Source & Target |

## Installation

### Build from source

```bash
go install github.com/vinaycharlie01/devsettings-mcp/cmd/server@latest
```

### Or clone and build

```bash
git clone https://github.com/vinaycharlie01/devsettings-mcp.git
cd devsettings-mcp
go build -o devsettings-mcp ./cmd/server
```

## MCP Configuration

### Claude Code / Claude Desktop

Add to your MCP settings:

```json
{
  "mcpServers": {
    "devsettings": {
      "command": "devsettings-mcp",
      "args": []
    }
  }
}
```

### VS Code

Add to `.vscode/mcp.json` or user settings:

```json
{
  "servers": {
    "devsettings": {
      "command": "devsettings-mcp",
      "args": []
    }
  }
}
```

### Zed

Add to `settings.json`:

```json
{
  "context_servers": {
    "devsettings": {
      "command": {
        "path": "devsettings-mcp"
      }
    }
  }
}
```

## MCP Tools

### `detect_editors`

Detect installed editors and their configuration paths.

```
No input required.
```

### `get_settings`

Read current settings for an editor.

```json
{
  "editor": "vscode",
  "scope": "user"
}
```

### `update_settings`

Update settings with deep merge support.

```json
{
  "editor": "vscode",
  "scope": "user",
  "settings": {
    "editor.formatOnSave": true,
    "editor.tabSize": 4
  },
  "merge_strategy": "merge"
}
```

### `create_backup`

Create a backup of editor configuration.

```json
{
  "editor": "vscode",
  "label": "before-go-setup"
}
```

### `restore_backup`

Restore a previous backup (supports dry-run).

```json
{
  "backup_id": "vscode-2026-11-10-143022",
  "dry_run": true
}
```

### `plan_migration`

Generate a migration plan between editors.

```json
{
  "source": "vscode",
  "target": "zed"
}
```

### `execute_migration`

Execute an approved migration plan.

```json
{
  "plan_id": "migration-vscode-zed-2026-11-10-143022"
}
```

## Example Conversation

```
User: "Set up my VS Code for Go development with format on save."

Agent calls: detect_editors → get_settings → update_settings
Result: VS Code configured with Go settings, backup created automatically.
```

```
User: "Migrate my VS Code settings to Zed."

Agent calls: plan_migration → (shows compatibility report) → execute_migration
Result: Compatible settings migrated, unsupported features listed.
```

## Development

### Prerequisites

- Go 1.24+
- [Task](https://taskfile.dev/) (optional, for task automation)

### Quick start

```bash
# Install dependencies
go mod download

# Run tests
go test ./... -v

# Build
go build -o devsettings-mcp ./cmd/server
```

### With Taskfile

```bash
task init        # Download dependencies
task fmt         # Format code
task test        # Run tests with race detection
task build       # Build binary
task ci          # Run full CI pipeline
```

## Architecture

```
cmd/server/main.go          — Entry point, wiring
internal/domain/            — Core types (Editor, Settings, Migration)
internal/ports/             — Interfaces (Editor, Normalizer)
internal/application/       — Business logic (Settings, Backup, Migration services)
internal/adapters/vscode/   — VS Code adapter + normalizer
internal/adapters/zed/      — Zed adapter + normalizer
internal/adapters/cursor/   — Cursor adapter
internal/mcp/               — MCP server and tool handlers
profiles/                   — Example development profiles
docs/                       — Research and documentation
```

## Security

- Read operations require no approval
- Write operations create automatic backups
- API keys, tokens, and passwords are redacted in backups
- Atomic file writes prevent partial corruption
- Path traversal is validated
- No arbitrary shell command execution

## License

Apache-2.0 — see [LICENSE](LICENSE) for details.
