package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
	"github.com/Gex-Development/gex-dev-kit/internal/ui/theme"
)

func TestWritePreviewPrintsPlainDiffWithoutANSI(t *testing.T) {
	grupos := []components.ChangeGroup{{
		Title: "Core",
		Changes: []components.Change{{
			Action: components.ChangeCreate,
			Path:   ".claude/skills/a.md",
			Detail: "2 linhas",
			Lines: []components.DiffLine{
				{Kind: components.DiffHunk, Text: "@@ -0,0 +1 @@"},
				{Kind: components.DiffAdded, Text: "linha nova"},
				{Kind: components.DiffRemoved, Text: "linha antiga"},
				{Kind: components.DiffContext, Text: "igual"},
			},
			Omitted: 3,
			Masked:  2,
		}},
	}}
	var out bytes.Buffer
	writePreview(&out, theme.New(theme.Options{Plain: true}), grupos)
	texto := out.String()
	for _, esperado := range []string{"Core", "criar", ".claude/skills/a.md", "+ linha nova", "- linha antiga", "  igual", "@@ -0,0 +1 @@", "… mais 3 linhas", "2 valores ocultos"} {
		if !strings.Contains(texto, esperado) {
			t.Fatalf("prévia sem %q: %q", esperado, texto)
		}
	}
	if strings.Contains(texto, "\x1b") {
		t.Fatalf("prévia com ANSI no tema Plain: %q", texto)
	}
}

func TestWritePreviewUsesSingularForOneOmittedLine(t *testing.T) {
	grupos := []components.ChangeGroup{{Title: "Core", Changes: []components.Change{{Action: components.ChangeUpdate, Path: "x", Omitted: 1}}}}
	var out bytes.Buffer
	writePreview(&out, theme.New(theme.Options{Plain: true}), grupos)
	if !strings.Contains(out.String(), "… mais 1 linha") || strings.Contains(out.String(), "linhas") {
		t.Fatalf("plural incorreto: %q", out.String())
	}
}

func TestWriteProgressWritesEachStepInPortuguese(t *testing.T) {
	var errOut bytes.Buffer
	progresso := writeProgress(&errOut, false)
	progresso(app.StepAgents)
	progresso(app.StepFiles)
	progresso(app.StepDone)
	esperado := "[..] Atualizando agentes\n[..] Gravando Core\n[ok] Core gravado\n"
	if errOut.String() != esperado {
		t.Fatalf("progresso: saída=%q esperado=%q", errOut.String(), esperado)
	}
}

func TestWriteProgressQuietWritesNothing(t *testing.T) {
	var errOut bytes.Buffer
	progresso := writeProgress(&errOut, true)
	progresso(app.StepAgents)
	progresso(app.StepFiles)
	progresso(app.StepDone)
	if errOut.Len() != 0 {
		t.Fatalf("modo silencioso escreveu: %q", errOut.String())
	}
}
