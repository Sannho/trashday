# AGENTS.md

Guidance for AI coding agents working on this repository.

## Project

TrashDay is a Windows-only disk cleaner: one Go binary (module `github.com/Sannho/trashday`) with a WebView2 UI ([go-webview2](https://github.com/jchv/go-webview2)) styled with Tailwind CSS.

| Path | Role |
|---|---|
| `cmd/trashday/main.go` | Window and JS bindings: `goI18n` → `ui.Texts`, `goTargets`, `goClean` (runs `cleaner.Clean` in a goroutine, reports back with `onProgress` / `onDone` through `Dispatch`) |
| `internal/cleaner/targets.go` | `Targets()`: every target's `ID`, `Group`, glob `Patterns`, `Quick`. `envPath` / `drivePath` validate env vars |
| `internal/cleaner/clean.go` | `Clean(target, dry)` / `cleanPath`: empty folders, keep roots, skip reparse points and `Content.Outlook`. `dry` = Analyze |
| `ui/ui.go` | Embeds the UI. `Page()` inlines CSS and JS into `index.html`; `Texts(locale)` loads `i18n/<lang>.json` with English fallback |
| `ui/index.html` | Markup only. Texts come from `[data-i18n]` / `[data-i18n-label]` keys |
| `ui/app.js` | UI behavior. `t(key, vars)` reads `ui` texts with `{placeholders}` |
| `ui/style.css` | Tailwind source with named component classes (`.card`, `.row`, `.btn-primary`…). Compiled to `ui/app.css` (generated, git-ignored, embedded) |
| `ui/i18n/<lang>.json` | `ui`, `groups`, `targets.<ID>.name` / `.desc`. English is the reference and the fallback |
| `build/windows/` | `trashday.manifest` (`highestAvailable`, common controls, DPI) and `trashday.ico` |
| `tools/icongen/` | `go run ./tools/icongen` regenerates the icon and `docs/images/icon.png` |
| `build.ps1` | Tailwind CLI (checksum-pinned) → CSS → `go test ./...` → `rsrc.syso` → `trashday.exe` |

## Commands

```powershell
.\build.ps1      # full build; required once before `go test` (generates ui/app.css)
go test ./...
go vet ./...
```

## Hard rules

- **Never run `trashday.exe` or call `cleaner.Clean(..., false)` on real paths.** It deletes files. Use `dry = true`, or tests on `t.TempDir()`.
- **Safety first.** Only add targets that the owning app or Windows rebuilds with no user loss. No history, cookies, sessions, settings, Recycle Bin, or anything that needs a service stopped. If unsure, don't add it, and say so.
- Build every path with `envPath` / `drivePath`. Never use `os.Getenv` directly in a pattern.
- Don't make `cleanPath` follow or remove symlinks or junctions, and don't make it delete a matched root.
- Every target `ID` and `Group` needs an entry in **every** `ui/i18n/*.json`: a `name` and an optional `desc` of at most 48 characters. `TestTranslations` enforces this.
- The saved selection is keyed by target `ID`. Renaming an ID resets that item for users.
- Never hard-code UI text in HTML or JS: add a key to every `ui/i18n/*.json`.
- Styles: reuse or add a named class in `ui/style.css` instead of repeating long utility lists.
- `rsrc.syso` must not stay in `cmd/trashday/`: it makes `go test` binaries require elevation. `build.ps1` creates it and removes it.
- Keep changes small, run `gofmt`, add tests for new logic, and note user-visible changes in `CHANGELOG.md` under **Unreleased**.
