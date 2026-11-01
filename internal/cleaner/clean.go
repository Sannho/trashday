package cleaner

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// Result sums what a cleaning run did. Failed counts entries that could not be removed,
// usually because a running program holds them open.
type Result struct {
	Freed  int64 `json:"freed"` // bytes; formatted by the page in the UI language
	Failed int   `json:"failed"`
}

// Add sums o into r.
func (r *Result) Add(o Result) { r.Freed += o.Freed; r.Failed += o.Failed }

// keepNames are folders never entered nor deleted, wherever they appear.
// Content.Outlook holds opened (and possibly edited) Outlook attachments.
var keepNames = map[string]bool{"content.outlook": true}

// Clean runs one target. Patterns with an empty path (rejected env var) are skipped.
// With dry set nothing is deleted and Freed is what a real run would free (locks aside).
// Overlapping matches (the same folder twice, or one inside another) are handled once,
// so Analyze never counts a file twice.
func Clean(t Target, dry bool) Result {
	if t.Run != nil {
		return t.Run(dry)
	}
	var matches []string
	for _, pattern := range t.Patterns {
		if pattern != "" {
			m, _ := filepath.Glob(pattern)
			matches = append(matches, m...)
		}
	}
	keys := make(map[string]string, len(matches)) // match -> lowercase key (Windows paths ignore case)
	for _, m := range matches {
		keys[m] = strings.ToLower(filepath.Clean(m))
	}
	sort.Slice(matches, func(i, j int) bool { return keys[matches[i]] < keys[matches[j]] }) // parents first

	var res Result
	var done []string
	for _, m := range matches {
		k := keys[m]
		if slices.ContainsFunc(done, func(d string) bool { return k == d || strings.HasPrefix(k, d+`\`) }) {
			continue
		}
		done = append(done, k)
		res.Add(cleanPath(m, dry))
	}
	return res
}

// cleanPath empties a directory but keeps the directory itself; a plain file is deleted.
// Symlinks, junctions and other reparse points are never followed nor removed: a cache
// the user moved to another drive with a junction stays as they set it up.
func cleanPath(path string, dry bool) Result {
	info, err := os.Lstat(path)
	if err != nil {
		return Result{}
	}
	if info.Mode().IsRegular() {
		if dry {
			return Result{Freed: info.Size()}
		}
		if os.Remove(path) != nil {
			return Result{Failed: 1}
		}
		return Result{Freed: info.Size()}
	}
	if !info.IsDir() {
		return Result{}
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Result{}
		}
		return Result{Failed: 1}
	}
	var res Result
	for _, e := range entries {
		if keepNames[strings.ToLower(e.Name())] {
			continue
		}
		p := filepath.Join(path, e.Name())
		res.Add(cleanPath(p, dry))
		if e.IsDir() && !dry {
			os.Remove(p) // fails silently if something inside was kept or locked
		}
	}
	return res
}
