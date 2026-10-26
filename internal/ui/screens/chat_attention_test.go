package screens

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

func TestDoctorLinesPutAttentionBeforeOkSoCollapsedCardShowsIt(t *testing.T) {
	report := app.Report{
		Items: []app.Status{
			{Path: ".agents/skills/a/SKILL.md", State: "ok"},
			{Path: ".agents/skills/b/SKILL.md", State: "ok"},
			{Path: ".claude/CLAUDE.md", State: "alterado"},
			{Path: ".codex/hooks.json", State: "pendente"},
		},
		Warnings: []app.Status{{Path: "agente claude", State: "aviso", Detail: "motivo"}},
	}
	lines := doctorLines(report, true, "/home/demo")
	if len(lines) != 5 {
		t.Fatalf("esperava 5 linhas, veio %d: %v", len(lines), lines)
	}
	for i, want := range []string{"alterado .claude/CLAUDE.md", "aviso · agente claude", "pendente .codex/hooks.json", "ok .agents/skills/a", "ok .agents/skills/b"} {
		if !strings.Contains(lines[i], want) {
			t.Fatalf("linha %d: %q não contém %q", i, lines[i], want)
		}
	}
}

func TestAgentBannerTitleNeverClaimsAbsentWhenBlocked(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change app.AgentChange
		want   string
	}{
		{"instalado pelo kit", app.AgentChange{Current: "2.1.306"}, "Claude Code v2.1.306"},
		{"bloqueado sem registro do kit", app.AgentChange{Blocked: true}, "Claude Code (não instalado pelo kit)"},
		{"bloqueado mas com versão", app.AgentChange{Current: "v2.0.0", Blocked: true}, "Claude Code v2.0.0"},
		{"ausente", app.AgentChange{}, "Claude Code (não instalado)"},
	} {
		if got := agentBannerTitle("Claude Code", tc.change); got != tc.want {
			t.Fatalf("%s: %q, esperado %q", tc.name, got, tc.want)
		}
	}
}

func TestChatSetupTwoBlockedAgentsEachKeepTheirOwnReason(t *testing.T) {
	d := chatInstallFakeDeps()
	m := chatSetupModel(d)
	m.agents = []app.AgentChange{
		{Name: "claude", Target: "2.1.0", Blocked: true, Reason: "caminhos do claude"},
		{Name: "codex", Target: "1.0.0", Blocked: true, Reason: "caminhos do codex"},
	}
	m = chatSetupQuestion(t, m, "Qual harness")
	m = chatSetupChoose(t, m, "claude", "codex")
	m = chatSetupQuestion(t, m, "Onde eu instalo?")
	text := chatTestTranscript(m)
	if !strings.Contains(text, "Claude Code: caminhos do claude") || !strings.Contains(text, "Codex: caminhos do codex") {
		t.Fatalf("cada agente precisa do seu motivo: %s", text)
	}
}
