package screens

import (
	"testing"
)

func TestCoverageMapIsComplete(t *testing.T) {
	paginas := map[string]bool{}
	for _, p := range showcasePages() {
		paginas[p.Name] = true
	}
	for i, c := range CoverageMap {
		if c.Lib == "" || c.Feature == "" || c.Page == "" {
			t.Fatalf("entrada %d sem Lib, Feature ou Page: %+v", i, c)
		}
		if !paginas[c.Page] {
			t.Fatalf("entrada %d (%s) aponta para página inexistente %q", i, c.Feature, c.Page)
		}
		if c.TestOnly {
			continue
		}
		if c.Demo == nil {
			t.Fatalf("entrada %d (%s) sem Demo", i, c.Feature)
		}
		if c.Marker == "" {
			t.Fatalf("entrada %d (%s) sem Marker", i, c.Feature)
		}
	}
}

func TestCoverageTestOnlyEhSoParaTesteDeTUI(t *testing.T) {
	for _, c := range CoverageMap {
		if !c.TestOnly {
			continue
		}
		if c.Lib != "x" || (c.Call != "teatest.NewTestModel" && c.Call != "golden.RequireEqual") {
			t.Fatalf("TestOnly indevido em %s (%s): só teatest e golden podem ser TestOnly", c.Feature, c.Call)
		}
	}
}

func TestCoverageCobreCadaLibDaTabela(t *testing.T) {
	libs := map[string]bool{}
	for _, c := range CoverageMap {
		libs[c.Lib] = true
	}
	for _, lib := range []string{"bubbletea", "bubbles", "huh", "lipgloss", "harmonica", "x"} {
		if !libs[lib] {
			t.Fatalf("lib %q da tabela §4 não aparece em CoverageMap", lib)
		}
	}
}
