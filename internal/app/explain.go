package app

import (
	"errors"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/githooks"
	"github.com/Gex-Development/gex-dev-kit/internal/kit"
)

func Explain(err error) Problem {
	if err == nil {
		return Problem{}
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
