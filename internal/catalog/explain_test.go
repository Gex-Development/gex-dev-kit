package catalog

import (
	"testing"
	"unicode/utf8"
)

func TestExplainCoversCatalogWithoutOrphans(t *testing.T) {
	known := map[string]bool{}
	for _, names := range [][]string{CoreSkills(), Skills(), Hooks(), GitHooks()} {
		for _, name := range names {
			known[name] = true
		}
	}
	for _, specialty := range Specialties() {
		known[specialty.Name] = true
	}
	for name := range known {
		item, ok := Explain(name)
		if !ok {
			t.Errorf("explicação ausente: %s", name)
			continue
		}
		for label, text := range map[string]string{"What": item.What, "UseCase": item.UseCase, "Analogy": item.Analogy} {
			if n := utf8.RuneCountInString(text); n < 40 || n > 400 {
				t.Errorf("%s %s tem %d caracteres (esperado 40 a 400)", name, label, n)
			}
		}
	}
	for name := range explanations {
		if !known[name] {
			t.Errorf("explicação sem item no catálogo: %s", name)
		}
	}
}
