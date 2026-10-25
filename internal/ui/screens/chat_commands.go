package screens

import (
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/ui/components"
)

var chatCommands = []components.PaletteItem{
	{Name: "setup", Description: "configurar skills, hooks e agentes"},
	{Name: "catalogo", Description: "ver skills e hooks com explicações"},
	{Name: "doctor", Description: "verificar a integridade da instalação"},
	{Name: "rollback", Description: "desfazer a última instalação"},
	{Name: "agentes", Description: "restaurar versões anteriores dos agentes"},
	{Name: "githooks", Description: "configurar git hooks de um repositório"},
	{Name: "addons", Description: "ver addons opcionais"},
	{Name: "menu", Description: "mostrar as opções como lista"},
	{Name: "voltar", Description: "voltar uma etapa"},
	{Name: "ajuda", Description: "comandos e atalhos"},
	{Name: "limpar", Description: "limpar a conversa"},
	{Name: "sair", Description: "encerrar"},
}

func chatIntentFor(text string) (ChatIntent, string, bool) {
	lower := strings.ToLower(strings.TrimSpace(text))
	if strings.HasPrefix(lower, "/") {
		switch lower {
		case "/setup":
			return IntentSetup, "", true
		case "/catalogo":
			return IntentCatalog, "", true
		case "/doctor":
			return IntentDoctor, "", true
		case "/rollback":
			return IntentRollback, "", true
		case "/agentes":
			return IntentRollbackAgents, "", true
		case "/githooks":
			return IntentGitHooks, "", true
		case "/addons":
			return IntentAddons, "", true
		case "/menu", "/voltar", "/ajuda", "/limpar", "/sair":
			return "", strings.TrimPrefix(lower, "/"), true
		}
		return "", "", false
	}
	for _, item := range []struct {
		words   []string
		intent  ChatIntent
		special string
	}{
		{[]string{"instal", "setup", "configur"}, IntentSetup, ""},
		{[]string{"catalog", "skill"}, IntentCatalog, ""},
		{[]string{"doctor", "integridade", "verific"}, IntentDoctor, ""},
		{[]string{"rollback", "desfaz"}, IntentRollback, ""},
		{[]string{"agente"}, IntentRollbackAgents, ""},
		{[]string{"git hook"}, IntentGitHooks, ""},
		{[]string{"addon"}, IntentAddons, ""},
		{[]string{"ajuda", "help"}, "", "ajuda"},
		{[]string{"voltar"}, "", "voltar"},
		{[]string{"menu"}, "", "menu"},
		{[]string{"sair", "tchau"}, "", "sair"},
	} {
		for _, word := range item.words {
			if strings.Contains(lower, word) {
				return item.intent, item.special, true
			}
		}
	}
	return "", "", false
}
