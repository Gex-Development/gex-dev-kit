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

func TestExplainErrosDeUsoMarcamUsage(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantTitle  string
		wantAction string
	}{
		{"skill desconhecida", errors.New("Catálogo: skill indisponível: inexistente"), "Skill desconhecida", "Veja `gex-dev-kit skills` para os nomes válidos."},
		{"hook de instalação desconhecido", errors.New("Catálogo: hook indisponível: inexistente"), "Hook desconhecido", "Veja `gex-dev-kit hooks`."},
		{"git hook desconhecido", errors.New("git hook desconhecido: post-merge"), "Hook desconhecido", "Veja `gex-dev-kit hooks`."},
		{"nenhum hook selecionado", errors.New("nenhum git hook selecionado"), "Hook não selecionado", "Informe um hook com `--hook NOME`. Veja `gex-dev-kit hooks`."},
		{"--run sem pre-push", errors.New("comandos de pre-push sem o hook pre-push selecionado"), "Opção sem efeito", "Adicione `--hook gex-pre-push` ou remova `--run`."},
		{"--pre-commit-run sem pre-commit", errors.New("comandos de pre-commit sem o hook pre-commit selecionado"), "Opção sem efeito", "Adicione `--hook gex-pre-commit` ou remova `--pre-commit-run`."},
		{"repositório ausente", githooks.ErrNotRepo, "Não é repositório Git", "Informe a raiz de um repositório Git (pasta com .git)."},
	}
	for _, c := range cases {
		got := Explain(c.err)
		if got.Title != c.wantTitle || got.Action != c.wantAction || !IsUsage(c.err) {
			t.Fatalf("%s: Explain = %+v, esperado título %q, ação %q e erro de uso", c.name, got, c.wantTitle, c.wantAction)
		}
		if strings.Contains(got.Cause, "Catálogo: ") {
			t.Fatalf("%s: causa carrega prefixo de camada: %q", c.name, got.Cause)
		}
	}
}

func TestExplainFalhaComumNaoEhUso(t *testing.T) {
	if IsUsage(errors.New("erro qualquer")) || IsUsage(nil) {
		t.Fatalf("falha comum marcada como erro de uso")
	}
}
