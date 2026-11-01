package ui

import (
	"strings"
	"testing"

	"github.com/Sannho/trashday/internal/cleaner"
)

// Every language file has every UI text, group and target of English; target IDs are
// unique (they key the saved selection); descriptions fit one line in the 440 px window.
func TestTranslations(t *testing.T) {
	en, err := Texts("en")
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := files.ReadDir("i18n")
	for _, f := range entries {
		lang := strings.TrimSuffix(f.Name(), ".json")
		m, err := Texts(lang)
		if err != nil || m["lang"] != lang {
			t.Fatalf("%s: %v (lang %v)", f.Name(), err, m["lang"])
		}
		for _, section := range []string{"ui", "groups"} {
			for k := range en[section].(map[string]any) {
				if _, ok := m[section].(map[string]any)[k]; !ok {
					t.Errorf("%s: missing %s.%s", f.Name(), section, k)
				}
			}
		}
		seen := map[string]bool{}
		for _, tg := range cleaner.Targets() {
			if seen[tg.ID] {
				t.Errorf("duplicate target ID %q", tg.ID)
			}
			seen[tg.ID] = true
			if _, ok := m["groups"].(map[string]any)[tg.Group]; !ok {
				t.Errorf("%s: missing groups.%s", f.Name(), tg.Group)
			}
			text, _ := m["targets"].(map[string]any)[tg.ID].(map[string]any)
			if name, _ := text["name"].(string); name == "" {
				t.Errorf("%s: missing targets.%s.name", f.Name(), tg.ID)
			}
			if desc, _ := text["desc"].(string); len([]rune(desc)) > 48 {
				t.Errorf("%s: targets.%s.desc too long (%d chars)", f.Name(), tg.ID, len([]rune(desc)))
			}
		}
	}
	if m, _ := Texts("de-DE"); m["lang"] != "en" {
		t.Error("unknown language should fall back to English")
	}
	if m, _ := Texts("fr-CA"); m["lang"] != "fr" {
		t.Error("fr-CA should pick French")
	}
}

func TestPageInlinesAssets(t *testing.T) {
	p := Page()
	if strings.Contains(p, `href="app.css"`) || strings.Contains(p, `src="app.js"`) {
		t.Error("stylesheet or script left as a link: the window cannot load it")
	}
}
