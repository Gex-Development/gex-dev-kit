package app

import (
	"errors"
	"slices"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func Explain(err error) Problem {
	if err == nil {
		return Problem{}
	}
	if p, ok := usageProblem(err); ok {
		return p
	}
	switch {
	case errors.Is(err, kit.ErrConflict):
		list := strings.TrimPrefix(err.Error(), kit.ErrConflict.Error()+": ")
		return Problem{
			Title:  "Rollback abortado",
			Cause:  "O rollback não alterou nada: estes arquivos mudaram depois da instalação: " + list + ".",
			Action: "Devolva-os à versão instalada (veja `doctor`) ou siga sem rollback; o kit nunca sobrescreve edição posterior.",
		}
	case errors.Is(err, kit.ErrChanged):
		return Problem{Title: "Prévia desatualizada", Cause: err.Error(), Action: "Rode o comando de novo para refazer a prévia."}
	case errors.Is(err, kit.ErrBusy):
		return Problem{Title: "Histórico ocupado", Cause: "Outra execução do kit pode estar ativa.", Action: "Aguarde. Se nada estiver rodando, veja `doctor` para o PID e o caminho do lock."}
	case errors.Is(err, kit.ErrIncomplete):
		return Problem{Title: "Instalação incompleta", Cause: err.Error(), Action: "Rode `gex-dev-kit rollback`."}
	case errors.Is(err, kit.ErrPruned):
		return Problem{
			Title:  "Limite do histórico",
			Cause:  "O rollback só alcança as 5 gravações mais recentes; o conteúdo desta gravação já saiu do histórico.",
			Action: "Restaure manualmente a partir de uma cópia ou reinstale com `install`.",
		}
	case errors.Is(err, kit.ErrSymlink):
		path := strings.TrimPrefix(err.Error(), kit.ErrSymlink.Error()+": ")
		return Problem{
			Title:  "Link simbólico recusado",
			Cause:  "`" + path + "` é link simbólico; o kit não segue links (proteção contra gravar fora do HOME).",
			Action: "Troque por arquivo ou pasta comum, ou teste com `--home`.",
		}
	case errors.Is(err, githooks.ErrUndocumented):
		return Problem{Title: "Comando não documentado", Cause: err.Error(), Action: "O comando precisa estar documentado no AGENTS.md, README, Makefile ou package.json do repositório."}
	case strings.Contains(err.Error(), "absoluto"):
		return Problem{Title: "Caminho inválido", Cause: err.Error(), Action: "Use caminho absoluto."}
	}
	return Problem{Title: "Falha inesperada", Cause: err.Error(), Action: "Rode `doctor` para ver o estado do kit."}
}

func IsUsage(err error) bool {
	_, ok := usageProblem(err)
	return ok
}

func usageProblem(err error) (Problem, bool) {
	if err == nil {
		return Problem{}, false
	}
	switch {
	case errors.Is(err, githooks.ErrNotRepo):
		return Problem{Title: "Não é repositório Git", Cause: "O caminho de --repo não está dentro de um repositório Git.", Action: "Informe a raiz de um repositório Git (pasta com .git)."}, true
	case errors.Is(err, catalog.ErrUnknownSkill):
		return Problem{Title: "Skill desconhecida", Cause: causaDe(err, catalog.ErrUnknownSkill), Action: "Veja `gex-dev-kit skills` para os nomes válidos."}, true
	case errors.Is(err, catalog.ErrUnknownHook):
		return Problem{Title: "Hook desconhecido", Cause: causaDe(err, catalog.ErrUnknownHook), Action: "Veja `gex-dev-kit hooks`."}, true
	case errors.Is(err, githooks.ErrUnknownHook):
		return Problem{Title: "Hook desconhecido", Cause: causaDe(err, githooks.ErrUnknownHook), Action: "Veja `gex-dev-kit hooks`."}, true
	case errors.Is(err, githooks.ErrNoHook):
		return Problem{Title: "Hook não selecionado", Cause: githooks.ErrNoHook.Error() + ".", Action: "Informe um hook com `--hook NOME`. Veja `gex-dev-kit hooks`."}, true
	case errors.Is(err, githooks.ErrPrePushWithoutHook):
		return Problem{Title: "Opção sem efeito", Cause: githooks.ErrPrePushWithoutHook.Error() + ".", Action: "Adicione `--hook gex-pre-push` ou remova `--run`."}, true
	case errors.Is(err, githooks.ErrPreCommitWithoutHook):
		return Problem{Title: "Opção sem efeito", Cause: githooks.ErrPreCommitWithoutHook.Error() + ".", Action: "Adicione `--hook gex-pre-commit` ou remova `--pre-commit-run`."}, true
	}
	return Problem{}, false
}

func causaDe(err error, sentinelas ...error) string {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if u := errors.Unwrap(e); u != nil && slices.ContainsFunc(sentinelas, func(s error) bool { return s == u }) {
			return e.Error()
		}
	}
	return err.Error()
}
