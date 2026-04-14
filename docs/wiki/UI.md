# UI Layer

The UI layer (`internal/adapters/ui/`) implements the user interface using the [Gio](https://gioui.org/) framework (see [ADR-0002](../adr/0002-record-ui-framework.md)).

## WindowManager

`views.WindowManager` is the root UI controller. It manages application state, user interactions, and rendering through Gio's immediate-mode rendering pipeline.

### Application States

| State | Description |
|-------|-------------|
| `StateSplash` | Animated splash screen on startup |
| `StateHome` | Main application view with sidebar navigation |

### Sidebar Tabs

The sidebar provides navigation between the main views:

1. **Dashboard** — most recent book and reading statistics
2. **Library** — grid view of all imported books
3. **Reading Sheets** — list of personal reading notes
4. **Reminders** — scheduled reading reminders

### Key Components

- **Book Grid** — responsive grid layout of imported books with status badges
- **Search** — live-filtering editor that filters the book library
- **Reader View** — page-by-page text reader for PDF/EPUB content
- **Sheet Detail View** — displays reading sheet with summary, quotes, and rating
- **Reminder View** — manages reading reminders with create/edit/delete

## Theme

The design system is defined in `internal/adapters/ui/theme/colors.go`:

| Constant | Hex | Description |
|----------|-----|-------------|
| `ColorVoidDark` | `#0D0D12` | Primary background (jet black) |
| `ColorSandGold` | `#F5A623` | Accent color (Egyptian gold) |
| `ColorCyberCyan` | `#2A3240` | Secondary surface (Nile slate) |
| `ColorGlassWhite` | `#F8F9FA` | Primary text (alabaster white) |
| `ColorPureBlack` | `#000000` | Deepest shadow |

### Utility Functions

- `hex2Color(c uint32) color.NRGBA` — converts a hex value to an NRGBA color
- `WithAlpha(c color.NRGBA, alpha uint8) color.NRGBA` — returns a color with modified alpha

## Platform Support

Gio provides native window management on:
- **Linux** — Wayland/X11 (requires `libwayland-dev`, `libxkbcommon-dev`)
- **macOS** — Cocoa
- **Windows** — Win32
