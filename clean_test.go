package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCleanPathEmptiesButKeepsRoot(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "a.tmp"), make([]byte, 10), 0o644)
	os.MkdirAll(filepath.Join(root, "sub", "deep"), 0o755)
	os.WriteFile(filepath.Join(root, "sub", "deep", "b.tmp"), make([]byte, 5), 0o644)

	res := cleanPath(root, false)

	if res.Freed != 15 || res.Failed != 0 {
		t.Fatalf("got %+v, want Freed=15 Failed=0", res)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("root should exist and be empty: %v %v", entries, err)
	}
}

func TestDryRunDeletesNothing(t *testing.T) {
	root := t.TempDir()
	f := filepath.Join(root, "sub", "a.tmp")
	os.MkdirAll(filepath.Dir(f), 0o755)
	os.WriteFile(f, make([]byte, 7), 0o644)

	res := cleanPath(root, true)

	if res.Freed != 7 {
		t.Errorf("got %+v, want Freed=7", res)
	}
	if _, err := os.Stat(f); err != nil {
		t.Fatalf("dry run deleted a file: %v", err)
	}
}

func TestEnvPathRejectsUnsafeValues(t *testing.T) {
	for _, v := range []string{"", `C:\`, `D:`, `relative\dir`, `C:\Users\a[b]c`, `C:\Users\*`} {
		t.Setenv("TRASHDAY_TEST_VAR", v)
		if p := envPath("TRASHDAY_TEST_VAR", "Temp"); p != "" {
			t.Errorf("envPath with %q = %q, want empty", v, p)
		}
	}
	t.Setenv("TRASHDAY_TEST_VAR", `C:\Users\me\AppData\Local`)
	if p := envPath("TRASHDAY_TEST_VAR", "Temp"); p != `C:\Users\me\AppData\Local\Temp` {
		t.Errorf("got %q", p)
	}
	if res := cleanTarget(Target{Patterns: []string{""}}, false); res != (Result{}) {
		t.Fatalf("got %+v, want nothing done", res)
	}
}

func TestDrivePath(t *testing.T) {
	t.Setenv("SystemDrive", "C:")
	if p := drivePath(`Users\*\Temp`); p != `C:\Users\*\Temp` {
		t.Errorf("got %q", p)
	}
	t.Setenv("SystemDrive", "")
	if p := drivePath("Users"); p != "" {
		t.Errorf("got %q, want empty", p)
	}
}

// junction creates a directory junction (no admin rights needed).
func junction(t *testing.T, link, target string) {
	t.Helper()
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("mklink: %v %s", err, out)
	}
}

func TestCleanPathNeverFollowsJunctions(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "outside")
	root := filepath.Join(dir, "root")
	os.MkdirAll(outside, 0o755)
	os.MkdirAll(root, 0o755)
	keep := filepath.Join(outside, "keep.txt")
	os.WriteFile(keep, []byte("x"), 0o644)
	junction(t, filepath.Join(root, "inner"), outside)
	moved := filepath.Join(dir, "moved-cache")
	junction(t, moved, outside)

	cleanPath(root, false)  // junction inside a target
	cleanPath(moved, false) // target that is itself a junction

	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("file behind junction was deleted: %v", err)
	}
	for _, link := range []string{filepath.Join(root, "inner"), moved} {
		if _, err := os.Lstat(link); err != nil {
			t.Errorf("junction %s was removed: %v", link, err)
		}
	}
}

func TestCleanPathKeepsOutlookAttachments(t *testing.T) {
	root := t.TempDir()
	att := filepath.Join(root, "Content.Outlook", "AB12", "edited.docx")
	os.MkdirAll(filepath.Dir(att), 0o755)
	os.WriteFile(att, []byte("x"), 0o644)
	os.WriteFile(filepath.Join(root, "cache.tmp"), []byte("x"), 0o644)

	cleanPath(root, false)

	if _, err := os.Stat(att); err != nil {
		t.Fatalf("Outlook attachment deleted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cache.tmp")); err == nil {
		t.Fatal("cache file not deleted")
	}
}

// Every target is translated, keys are unique (they key the saved selection),
// and descriptions stay short enough for one line in the 440 px window.
func TestTargetTexts(t *testing.T) {
	seen := map[string]bool{}
	for _, tg := range targets() {
		for _, s := range []tr{tg.Group, tg.Name} {
			if s.FR == "" || s.EN == "" {
				t.Errorf("missing translation: %+v", s)
			}
		}
		if (tg.Desc.FR == "") != (tg.Desc.EN == "") {
			t.Errorf("description in one language only: %+v", tg.Desc)
		}
		for _, d := range []string{tg.Desc.FR, tg.Desc.EN} {
			if n := len([]rune(d)); n > 48 {
				t.Errorf("description too long (%d chars): %q", n, d)
			}
		}
		if seen[tg.Name.EN] {
			t.Errorf("duplicate key %q", tg.Name.EN)
		}
		seen[tg.Name.EN] = true
	}
	if (tr{"a", "b"}).in("fr") != "a" || (tr{"a", "b"}).in("de") != "b" {
		t.Error("tr.in should pick French for fr, English otherwise")
	}
}
