# Changelog

All notable changes are listed here. Versions follow [Semantic Versioning](https://semver.org).

## [Unreleased]

### Changed
- Project layout: `cmd/trashday`, `internal/cleaner`, `ui/`, `build/windows`, `tools/icongen`.
- UI split into `ui/index.html`, `ui/app.js` and `ui/style.css`, with texts in `ui/i18n/<lang>.json`. Add a language by adding a file.
- "Current user temp" and "All users temp" are merged into **Users temp**.
- Saved selections are now keyed by target ID: items checked in 1.0.0 need to be checked again once.

### Fixed
- **Analyze** counted some files twice: the current user's Temp in both Temp items, and Chromium's `GrShaderCache\GPUCache` / `ShaderCache\GPUCache`. Overlapping folders are now counted once.

## [1.0.0] - 2026-11-01

First release.

### Added
- **Clean**, **Analyze** (dry run) and **Quick clean** actions.
- Targets: Windows temp, Prefetch, error reports, crash dumps, logs, Store apps, thumbnails, temporary Internet files; Firefox, Chrome, Edge, Brave and Opera caches; shader, Steam, Epic, Discord, VS Code and Java caches; Go, npm, Yarn, pip and NuGet caches.
- Safety: validated paths, no symlink or junction following, Outlook attachments kept, files in use skipped.
- Light and dark themes, English and French.

[Unreleased]: https://github.com/Sannho/trashday/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Sannho/trashday/releases/tag/v1.0.0
