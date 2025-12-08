# dbsurf

A terminal UI for managing and connecting to databases. Built with Go and Bubble Tea. Uses VIM motions.

## Tech Stack

- Go
- Bubble Tea (TUI framework)
- Lip Gloss (styling)
- Bubbles (text input components)

## Project Structure

```
dbsurf/
├── main.go           # Entry point, initializes Bubble Tea program
├── app/
│   ├── model.go      # App struct, New(), styles, constants
│   ├── app.go        # Bubble Tea interface (Init, Update, View) + renderFrame
│   ├── list.go       # List view (updateList, viewList)
│   └── input.go      # Input view (updateInput, viewInput)
├── config/
│   └── config.go     # Connection storage (~/.config/dbsurf/config.json)
└── db/
    └── db.go         # Database operations (connect, query, list databases)
```

## Architecture

The app uses Bubble Tea's Elm-style architecture:
- `*App` uses pointer receivers for all methods
- Mode-based routing in `Update()` and `View()` switches between views
- Each view has its own `update*` and `view*` methods in separate files

## Commands

- `go run .` - Run the app
- `go build .` - Build binary

## Config

Connections stored at `~/.config/dbsurf/config.json`

## Supported Databases

- PostgreSQL
- MySQL
