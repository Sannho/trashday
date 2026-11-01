# Contributing to TrashDay

Thanks for helping! TrashDay has one promise: **clean up the PC safely**. Every change is judged against it.

## The safety rule

A target may only contain data that its owner rebuilds by itself, with no loss for the user:

- ✅ caches, temp files, crash dumps, logs, thumbnails
- ❌ history, cookies, sessions, settings, the Recycle Bin, downloads, documents
- ❌ anything that needs a service stopped, breaks offline use, or can't be undone (DISM, `Windows.old`, `Installer`, restore points)
- ❌ registry "cleaning"

When in doubt, leave it out. Open an issue to discuss it first.

## Adding a target

Targets live in [`targets.go`](targets.go). Each one needs:

1. **Paths** built with `envPath(...)` or `drivePath(...)`, never a raw environment variable. These helpers reject empty, relative, drive-root and glob-like values.
2. **Texts in English and French**: the group, the name, and an optional one-line description of at most 48 characters.
3. **A reason** in the pull request: what the folder holds, who rebuilds it, and a source if you have one.

`cleanPath` empties folders but keeps them, never follows symlinks or junctions, and always skips `Content.Outlook`. Don't weaken these rules. If you change them, add a test.

## Build and test

You need Windows and [Go](https://go.dev/dl/) 1.26+.

```powershell
.\build.ps1      # Tailwind CSS, tests, then trashday.exe
go test ./...    # tests only (run build.ps1 once first: it generates ui.css)
```

- Use **Analyze** to try a change: it measures without deleting anything.
- Only press **Clean** on a machine where you accept losing the selected caches.

## Pull requests

- Keep each PR small and focused, and explain why.
- Run `gofmt`, then make sure `.\build.ps1` passes. GitHub Actions runs the same build on every PR.
- UI changes: edit [`index.html`](index.html) with Tailwind classes only. Add every new text in both languages, and include a screenshot.

## License

By contributing, you agree that your contribution is licensed under the [GPL-3.0](LICENSE), like the rest of the project.

## Releases

Maintainers push a tag such as `v1.2.0`. The [Build workflow](.github/workflows/build.yml) then publishes `trashday.exe` and its SHA-256 to a GitHub release.
