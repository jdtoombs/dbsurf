# dbsurf

A terminal UI for managing and connecting to databases. Built with Go and Bubble Tea. Uses VIM motions.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install jdtoombs/tap/dbsurf
```

### Debian/Ubuntu

```bash
curl -s https://packagecloud.io/install/repositories/jonny-toombs/dbsurf/script.deb.sh | sudo bash
sudo apt-get install dbsurf
```

## Current Features

- **Connection Management**: Save, delete, and manage multiple database connections with automatic database type detection
- **VIM Navigation**: Navigate lists and results using `j/k` (up/down), `h/l` (left/right for row navigation)
- **Database Browser**: Browse and select databases on connected servers (MySQL, SQL Server)
- **Table Browser**: Quick access to tables via `ctrl+t`, with search filtering
- **Query Execution**: Run custom SQL queries with results displayed in a JSON-like format
- **Inline Editing**: Edit field values directly with `i` key, generates UPDATE statements with confirmation
- **Search/Filter**: Filter databases, tables, and query results with `/` key
- **Copy to Clipboard**: Copy current record as JSON with `ctrl+c`
- **Primary Key Detection**: Automatic PK detection for safe UPDATE generation

## Tech Stack

- Go
- Bubble Tea (TUI framework)
- Lip Gloss (styling)
- Bubbles (text input components)

## Architecture

The app uses Bubble Tea's Elm-style architecture:
- `*App` uses pointer receivers for all methods
- Mode-based routing in `Update()` and `View()` switches between views
- Each view has its own `update*` and `view*` methods in separate files

## Config

Connections stored at `~/.config/dbsurf/config.json`

## Supported Databases

- PostgreSQL
- MySQL
- SQL Server
