package catalog

type Agent int

const (
	AgentClaude Agent = iota
	AgentCodex
)

type HookScope int

const (
	HookScopeAgent HookScope = iota
	HookScopeGit
)

type HookSupport struct {
	Supported bool
	Event     string
	Matcher   string
	Decision  string
	Reason    string
	Trust     bool
}

type HookSpec struct {
	Name        string
	Command     string
	Description string
	Scope       HookScope
	Policy      string
	Claude      HookSupport
	Codex       HookSupport
	GitHook     string
	Source      string
}

func HookSpecs() []HookSpec {
	gitReason := "Hook do Git, não de agente: roda pelo próprio Git, não pelo Claude Code nem pelo Codex."
	return []HookSpec{
		{
			Name:        "gex-conventional-commits",
			Command:     "commit-msg",
			Description: "Bloqueia commit com mensagem fora do padrão Conventional Commits. Só age em repositório GEX.",
			Scope:       HookScopeAgent,
			Policy:      "gex",
			Claude:      HookSupport{Supported: true, Event: "PreToolUse", Matcher: "Bash", Decision: "bloqueia"},
			Codex:       HookSupport{Supported: true, Event: "PreToolUse", Matcher: "Bash", Decision: "bloqueia", Trust: true},
			Source:      "Autoria GEX. Eventos: https://code.claude.com/docs/en/hooks e https://learn.chatgpt.com/docs/hooks",
		},
		{
			Name:        "gex-sensitive-action-warning",
			Command:     "sensitive-action",
			Description: "Pede confirmação antes de ação sensível, como publicar, apagar ou ler segredo. Só age em repositório GEX.",
			Scope:       HookScopeAgent,
			Policy:      "gex",
			Claude:      HookSupport{Supported: true, Event: "PreToolUse", Matcher: "Bash|PowerShell|Read", Decision: "pergunta"},
			Codex:       HookSupport{Reason: "O Codex ainda não suporta ask: o hook seria marcado como falho e a ação continuaria. Deny bloquearia uma ação autorizada sem saída."},
			Source:      "Autoria GEX. Evento: https://code.claude.com/docs/en/hooks",
		},
		{
			Name:        "gex-pre-pr",
			Command:     "pre-pr",
			Description: "Pede confirmação na abertura de PR e lembra o checklist do preflight. Só age em repositório GEX.",
			Scope:       HookScopeAgent,
			Policy:      "gex",
			Claude:      HookSupport{Supported: true, Event: "PreToolUse", Matcher: "Bash|PowerShell", Decision: "pergunta"},
			Codex:       HookSupport{Reason: "Mesmo motivo do ask: o Codex ainda não suporta pergunta antes da ação."},
			Source:      "Autoria GEX. Evento: https://code.claude.com/docs/en/hooks",
		},
		{
			Name:        "gex-session-start",
			Command:     "session-start",
			Description: "Informa no início da sessão se o diretório é repositório GEX, quais skills Core estão instaladas e se o bloco Core existe. Opt-in; só informa.",
			Scope:       HookScopeAgent,
			Policy:      "todos",
			Claude:      HookSupport{Supported: true, Event: "SessionStart", Matcher: "startup|resume|clear|compact|fork", Decision: "contexto"},
			Codex:       HookSupport{Supported: true, Event: "SessionStart", Matcher: "startup|resume|clear|compact|fork", Decision: "contexto", Trust: true},
			Source:      "Autoria GEX. Eventos: https://code.claude.com/docs/en/hooks e https://learn.chatgpt.com/docs/hooks",
		},
		{
			Name:        "gex-commit-msg",
			Command:     "commit-msg",
			Description: "Valida a mensagem do commit feito pelo Git com o padrão Conventional Commits. Só é gravado com --repo e confirmação.",
			Scope:       HookScopeGit,
			Policy:      "todos",
			Claude:      HookSupport{Reason: gitReason},
			Codex:       HookSupport{Reason: gitReason},
			GitHook:     "commit-msg",
			Source:      "Autoria GEX. Contrato do Git: https://git-scm.com/docs/githooks",
		},
		{
			Name:        "gex-pre-commit",
			Command:     "pre-commit",
			Description: "Bloqueia commit com possível credencial no diff e roda os comandos rápidos documentados pelo projeto.",
			Scope:       HookScopeGit,
			Policy:      "todos",
			Claude:      HookSupport{Reason: gitReason},
			Codex:       HookSupport{Reason: gitReason},
			GitHook:     "pre-commit",
			Source:      "Autoria GEX. Contrato do Git: https://git-scm.com/docs/githooks",
		},
		{
			Name:        "gex-pre-push",
			Description: "Roda antes do push os comandos documentados pelo projeto, passados com --run. Só é gravado com --repo e confirmação.",
			Scope:       HookScopeGit,
			Policy:      "todos",
			Claude:      HookSupport{Reason: gitReason},
			Codex:       HookSupport{Reason: gitReason},
			GitHook:     "pre-push",
			Source:      "Autoria GEX. Contrato do Git: https://git-scm.com/docs/githooks",
		},
	}
}

func HookSpecByName(name string) (HookSpec, bool) {
	for _, spec := range HookSpecs() {
		if spec.Name == name {
			return spec, true
		}
	}
	return HookSpec{}, false
}

func GitHooks() []string {
	var names []string
	for _, spec := range HookSpecs() {
		if spec.Scope == HookScopeGit {
			names = append(names, spec.Name)
		}
	}
	return names
}
