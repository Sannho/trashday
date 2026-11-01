# Contributing to TrashDay

Thanks for helping! TrashDay has one promise: **clean up the PC safely**. Every change is judged against it.

## The safety rule

A target may only contain data that its owner rebuilds by itself, with no loss for the user:

- ✅ caches, temp files, crash dumps, logs, thumbnails
- ❌ history, cookies, sessions, settings, downloads, documents (the Recycle Bin item is the one exception: off by default, never in Quick clean)
- ❌ anything that needs a service stopped, breaks offline use, or can't be undone (DISM, `Windows.old`, `Installer`, restore points)
- ❌ registry "cleaning"

When in doubt, leave it out. Open a [New target](https://github.com/Sannho/trashday/issues/new?template=new_target.yml) issue to discuss it first.

## Project layout

| Path | What lives there |
|---|---|
| `cmd/trashday/` | Entry point: the WebView2 window and the bindings between the UI and the cleaner |
| `internal/cleaner/` | `targets.go`: every target (ID, group, paths). `clean.go`: the safe delete logic. Tests |
| `ui/` | `index.html` (markup), `app.js` (behavior), `style.css` (Tailwind source), `ui.go` (embedding) |
| `ui/i18n/` | One JSON file per language |
| `build/windows/` | Manifest and icon embedded in the exe |
| `tools/icongen/` | Draws the icon: `go run ./tools/icongen` |

## Adding a target

1. Add it to [`internal/cleaner/targets.go`](internal/cleaner/targets.go) with a unique `ID` and an existing `Group`.
2. Build every path with `envPath(...)` or `drivePath(...)`, never a raw environment variable. These helpers reject empty, relative, drive-root and glob-like values.
3. Add its `name`, and an optional `desc` of at most 48 characters, under `targets.<ID>` in **every** [`ui/i18n/*.json`](ui/i18n).
4. In the pull request, explain what the folder holds and who rebuilds it, with a source if you have one.

`cleanPath` empties folders but keeps them, never follows symlinks or junctions, and always skips `Content.Outlook`. Don't weaken these rules. If you change them, add a test.

## Adding a language

Copy [`ui/i18n/en.json`](ui/i18n/en.json) to `ui/i18n/<code>.json`, for example `de.json`, and translate the values. TrashDay picks it automatically when Windows uses that language. The tests check that no key is missing.

## Build and test

You need Windows and [Go](https://go.dev/dl/) 1.26+.

```powershell
.\build.ps1      # Tailwind CSS, tests, then trashday.exe
go test ./...    # tests only (run build.ps1 once first: it generates ui/app.css)
```

- Use **Analyze** to try a change: it measures without deleting anything.
- Only press **Clean** on a machine where you accept losing the selected caches.

## Pull requests

- Keep each PR small and focused, and explain why.
- Run `gofmt`, then make sure `.\build.ps1` passes. It is the same build that GitHub Actions runs for releases.
- UI changes:
  - Put texts in `ui/i18n`, never in HTML or JS.
  - Put repeated styles in a named class in `ui/style.css`.
  - Include a screenshot.
- Add a line to [`CHANGELOG.md`](CHANGELOG.md) under **Unreleased**.

## License

By contributing, you agree that your contribution is licensed under the [GPL-3.0](LICENSE), like the rest of the project.

## Releases

Maintainers move the **Unreleased** changelog entries under a new version, then publish a release on GitHub (Releases → Draft a new release, with a tag such as `v1.2.0`). The [Build workflow](.github/workflows/build.yml) then tests, builds and attaches `trashday.exe` and its SHA-256 to that release. You can also run the workflow by hand from the Actions tab: the exe is then kept as a run artifact.
