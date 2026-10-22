package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func TestExplainTranslatesKitSentinels(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantTitle  string
		wantCause  string
		wantAction string
	}{
		{
			"conflito lista os arquivos",
			fmt.Errorf("%w: .claude/CLAUDE.md, .codex/AGENTS.md", kit.ErrConflict),
			"Rollback abortado",
			"O rollback não alterou nada: estes arquivos mudaram depois da instalação: .claude/CLAUDE.md, .codex/AGENTS.md.",
			"Devolva-os à versão instalada (veja `doctor`) ou siga sem rollback; o kit nunca sobrescreve edição posterior.",
		},
		{
			"prévia desatualizada",
			fmt.Errorf("%w: .claude/CLAUDE.md", kit.ErrChanged),
			"Prévia desatualizada",
			"arquivo alterado desde a prévia: .claude/CLAUDE.md",
			"Rode o comando de novo para refazer a prévia.",
		},
		{
			"histórico ocupado",
			fmt.Errorf("%w: arquivo existe", kit.ErrBusy),
			"Histórico ocupado",
			"Outra execução do kit pode estar ativa.",
			"Aguarde. Se nada estiver rodando, veja `doctor` para o PID e o caminho do lock.",
		},
		{
			"instalação incompleta",
			kit.ErrIncomplete,
			"Instalação incompleta",
			kit.ErrIncomplete.Error(),
			"Rode `gex-dev-kit rollback`.",
		},
		{
			"link simbólico",
			fmt.Errorf("%w: .claude", kit.ErrSymlink),
			"Link simbólico recusado",
			"`.claude` é link simbólico; o kit não segue links (proteção contra gravar fora do HOME).",
			"Troque por arquivo ou pasta comum, ou teste com `--home`.",
		},
		{
			"limite do histórico",
			fmt.Errorf("%w; restaure manualmente", kit.ErrPruned),
			"Limite do histórico",
			"O rollback só alcança as 5 gravações mais recentes; o conteúdo desta gravação já saiu do histórico.",
			"Restaure manualmente a partir de uma cópia ou reinstale com `install`.",
		},
		{
			"caminho relativo",
			errors.New("HOME deve ser absoluto e normalizado"),
			"Caminho inválido",
			"HOME deve ser absoluto e normalizado",
			"Use caminho absoluto.",
		},
		{
			"erro desconhecido",
			errors.New("erro qualquer"),
			"Falha inesperada",
			"erro qualquer",
			"Rode `doctor` para ver o estado do kit.",
		},
	}
	for _, c := range cases {
		got := Explain(c.err)
		if got.Title != c.wantTitle || got.Cause != c.wantCause || got.Action != c.wantAction {
			t.Fatalf("%s: Explain = %+v, esperado título %q, causa %q, ação %q", c.name, got, c.wantTitle, c.wantCause, c.wantAction)
		}
	}
}

func TestExplainNilIsEmpty(t *testing.T) {
	if got := Explain(nil); got != (Problem{}) {
		t.Fatalf("Explain(nil) = %+v, esperado vazio", got)
	}
}

func TestExplainConflictListsEveryFile(t *testing.T) {
	got := Explain(fmt.Errorf("%w: a.md, b.md", kit.ErrConflict))
	if !strings.Contains(got.Cause, "a.md, b.md") {
		t.Fatalf("causa do conflito não lista os arquivos: %q", got.Cause)
	}
}

func TestExplainUndocumentedCommand(t *testing.T) {
	p := Explain(fmt.Errorf("%w: make test", githooks.ErrUndocumented))
	if p.Title != "Comando não documentado" || !strings.Contains(p.Action, "precisa estar documentado") {
		t.Fatalf("explicação: %+v", p)
	}
}
