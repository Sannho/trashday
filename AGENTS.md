# AGENTS.md

Guidance for AI coding agents working on this repository.

## Project

TrashDay is a Windows-only disk cleaner: one Go binary with a WebView2 UI ([go-webview2](https://github.com/jchv/go-webview2)) and Tailwind CSS.

| File | Role |
|---|---|
| `targets.go` | All cleanable targets: glob paths and English/French texts. `envPath` / `drivePath` validate env vars |
| `clean.go` | `cleanTarget` / `cleanPath`: empty folders, keep roots, skip reparse points and `Content.Outlook`. `dry` = Analyze |
| `main.go` | Window, JS bindings (`goTargets`, `goClean`), cleaning in a goroutine, results via `Dispatch` |
| `index.html` | Entire UI and its i18n strings (`T.fr` / `T.en`). Tailwind classes only |
| `ui.input.css` | Tailwind entry. `ui.css` is generated, git-ignored, and embedded at build |
| `build.ps1` | Fetches the Tailwind CLI (checksum-pinned), builds CSS, runs tests, embeds manifest and icon, builds the exe |
| `icon_gen.go` | `go run icon_gen.go` regenerates `trashday.ico` (build-ignored) |

## Commands

```powershell
.\build.ps1      # full build; required once before `go test` (generates ui.css)
go test ./...
go vet ./...
```

## Hard rules

- **Never run `trashday.exe` or call `cleanTarget(..., false)` on real paths.** It deletes files. Use `dry = true`, or tests on `t.TempDir()`.
- **Safety first.** Only add targets that the owning app or Windows rebuilds with no user loss. No history, cookies, sessions, settings, Recycle Bin, or anything that needs a service stopped. If unsure, don't add it, and say so.
- Build every path with `envPath` / `drivePath`. Never use `os.Getenv` directly in a pattern.
- Don't make `cleanPath` follow or remove symlinks or junctions, and don't make it delete a matched root.
- Every target needs `Name` and `Group` in both `FR` and `EN`. An optional `Desc` must be at most 48 characters. `TestTargetTexts` enforces this.
- The saved selection is keyed by `Name.EN`. Renaming it resets that item for users.
- Any UI text goes in both `T.fr` and `T.en` in `index.html`.
- Bindings run on the UI thread: long work goes in a goroutine and reports back through `call(...)`, which wraps `Dispatch`.
- `rsrc.syso` must not stay in the tree: it makes `go test` binaries require elevation. `build.ps1` creates it and removes it.
- Keep changes small, run `gofmt`, and add a test for new logic in `clean.go` or `targets.go`.
