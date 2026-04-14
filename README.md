# Orus

> A unified desktop reading environment for PDF and EPUB formats, with advanced reading tracking, personal reading sheets, scheduled reminders, and export capabilities — wrapped in a glassy Mecha-Egyptian aesthetic.

[![CI](https://github.com/MiltonJ23/Orus/actions/workflows/ci.yml/badge.svg)](https://github.com/MiltonJ23/Orus/actions/workflows/ci.yml)
[![Release](https://github.com/MiltonJ23/Orus/actions/workflows/release.yml/badge.svg)](https://github.com/MiltonJ23/Orus/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg)](https://go.dev/)

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Build & Run](#build--run)
- [Testing](#testing)
- [Configuration](#configuration)
- [Export Formats](#export-formats)
- [Design System](#design-system)
- [Architecture Decision Records](#architecture-decision-records)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Orus is a cross-platform desktop application built in Go with the [Gio UI](https://gioui.org/) framework. It provides a self-contained reading environment for PDF and EPUB files, local-first and privacy-respecting. All data is stored in a single SQLite database.

### Why Orus?

- **Unified reader** — PDF and EPUB in one window, same interface.
- **Reading tracker** — automatic session tracking with per-book progress.
- **Reading sheets** — personal notes, ratings, quotes, and tags per book.
- **Scheduled reminders** — configurable reading reminders (daily, weekly, weekdays, once).
- **Library export** — share your library as JSON, Markdown, or plain text.
- **Cross-platform** — runs on Linux, macOS, and Windows.

---

## Features

| Feature | Description |
|---------|-------------|
| **PDF/EPUB Import** | Import books from local files with automatic metadata extraction |
| **In-App Reader** | Read directly within Orus with page-by-page navigation |
| **Reading Sessions** | Automatic tracking of reading position and time |
| **Reading Sheets** | Personal notes: summary, quotes, rating (★), and tags |
| **Reminders** | Scheduled reading reminders with multiple frequencies |
| **Library Export** | Export to JSON, Markdown, or plain text |
| **Search** | Live search/filter across your library |
| **Bookmarks** | Annotate pages with bookmarks and highlights |

---

## Architecture

Orus follows **Clean Architecture** with strict layer separation:

```
┌──────────────────────────────────────────────┐
│                  cmd/orus/                    │  Entry point
├──────────────────────────────────────────────┤
│              internal/adapters/               │  Infrastructure
│  ┌──────────┐ ┌────────┐ ┌────────────────┐  │
│  │    UI    │ │Storage │ │   Extractor    │  │
│  │  (Gio)  │ │(SQLite)│ │  (PDF/EPUB)    │  │
│  └────┬─────┘ └───┬────┘ └──────┬─────────┘  │
├───────┼───────────┼─────────────┼────────────┤
│       └───────────┼─────────────┘            │
│              internal/port/                   │  Interfaces
│         (Repository, ContentReader,           │
│          MetadataExtractor, Notifier)         │
├──────────────────────────────────────────────┤
│             internal/service/                 │  Application Logic
│   (Library, Tracker, ReadingSheet,            │
│    Reminder, Sharing)                         │
├──────────────────────────────────────────────┤
│             internal/domain/                  │  Business Entities
│   (Book, ReadingSession, Annotation,          │
│    ReadingSheet, Reminder)                    │
└──────────────────────────────────────────────┘
```

**Dependency rule**: outer layers depend on inner layers, never the reverse. All cross-layer communication goes through interfaces defined in `internal/port/`.

---

## Project Structure

```
Orus/
├── cmd/
│   └── orus/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Business entities and validation
│   │   ├── book.go                 # Book entity
│   │   ├── session.go              # ReadingSession entity
│   │   ├── annotation.go           # Bookmark/Highlight entity
│   │   ├── reading_sheet.go        # ReadingSheet entity
│   │   └── reminder.go             # Reminder entity
│   ├── port/                       # Interface definitions (ports)
│   │   ├── repository.go           # Book, Session, Annotation repos
│   │   ├── extended_repository.go  # ReadingSheet, Reminder repos, Notifier
│   │   ├── content_reader.go       # Text extraction interface
│   │   └── file_system.go          # Metadata extraction interface
│   ├── service/                    # Application services
│   │   ├── library.go              # Book import and library management
│   │   ├── tracker.go              # Reading progress tracking
│   │   ├── reading_sheet_service.go# Reading sheet CRUD
│   │   ├── reminder_service.go     # Reminder scheduling
│   │   └── sharing_service.go      # Export to JSON/Markdown/Text
│   └── adapters/                   # Infrastructure implementations
│       ├── extractor/              # PDF/EPUB metadata and text extraction
│       ├── notifier/               # Notification system (log-based)
│       ├── storage/sqlite/         # SQLite persistence layer
│       └── ui/                     # Gio UI components
│           ├── theme/              # Design system (colors)
│           └── views/              # Window manager and views
├── docs/
│   ├── adr/                        # Architecture Decision Records
│   └── wiki/                       # Project wiki documentation
├── .github/
│   └── workflows/                  # CI/CD pipelines
├── Makefile                        # Build automation
├── go.mod                          # Go module definition
├── CONTRIBUTING.md                 # Contribution guidelines
└── LICENSE                         # Apache 2.0
```

---

## Prerequisites

- **Go** >= 1.25
- **Platform dependencies** (for Gio UI):
  - **Linux**: `libwayland-dev`, `libxkbcommon-dev`, `libgles2-mesa-dev`
  - **macOS**: Xcode Command Line Tools
  - **Windows**: no additional dependencies

### Linux (Debian/Ubuntu)

```bash
sudo apt-get install -y libwayland-dev libxkbcommon-dev libgles2-mesa-dev
```

### macOS

```bash
xcode-select --install
```

---

## Getting Started

```bash
# Clone the repository
git clone https://github.com/MiltonJ23/Orus.git
cd Orus

# Download dependencies
go mod download

# Build and run
make run
```

---

## Build & Run

```bash
# Build the binary
make build

# Run the application
make run

# Clean build artifacts
make clean

# Format code
make fmt
```

The binary is written to `bin/orus`.

---

## Testing

```bash
# Run all tests
make test

# Run tests with coverage report
make test-with-coverage

# View coverage report
make consult-coverage
```

---

## Configuration

Orus stores its database in the current working directory as `orus.db`. The SQLite database is created automatically on first run.

### Database Schema

| Table | Purpose |
|-------|---------|
| `books` | Imported book metadata |
| `sessions` | Reading session tracking |
| `annotations` | Bookmarks and highlights |
| `reading_sheets` | Personal reading notes |
| `reminders` | Scheduled reading reminders |

---

## Export Formats

Orus supports exporting your library in three formats:

| Format | Extension | Description |
|--------|-----------|-------------|
| **Markdown** | `.md` | Formatted with headings, stars, and blockquotes |
| **JSON** | `.json` | Structured data for programmatic consumption |
| **Plain Text** | `.txt` | Simple text format |

Export uses the OS-native folder picker (zenity/kdialog on Linux, osascript on macOS, PowerShell on Windows).

---

## Design System

Orus uses a **Mecha-Egyptian** aesthetic with the following color palette:

| Name | Hex | Usage |
|------|-----|-------|
| Void Dark | `#0D0D12` | Primary background |
| Sand Gold | `#F5A623` | Accent, highlights |
| Cyber Cyan | `#2A3240` | Secondary surfaces |
| Glass White | `#F8F9FA` | Text on dark backgrounds |
| Pure Black | `#000000` | Deepest shadow |

Theme constants are defined in `internal/adapters/ui/theme/colors.go`.

---

## Architecture Decision Records

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-0001](docs/adr/0001-record-architecture-pattern.md) | Clean Architecture | Accepted |
| [ADR-0002](docs/adr/0002-record-ui-framework.md) | Gio UI Framework | Accepted |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full contribution guide, including:

- Branching strategy (`feat/*` → `dev` → `stage` → `main`)
- Semantic commit conventions
- PR checklist and review process
- Visual fidelity requirements

---

## License

This project is licensed under the **Apache License 2.0** — see the [LICENSE](LICENSE) file for details.
