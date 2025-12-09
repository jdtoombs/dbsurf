# dbsurf

A terminal UI for managing and connecting to databases. Built with Go and Bubble Tea. Uses VIM motions.

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
