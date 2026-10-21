package catalog

type CatalogReference struct {
	Name   string
	Status string
	Reason string
	Source string
}

func ReferencedSkills() []CatalogReference {
	const matt = "https://github.com/mattpocock/skills/tree/" + MattRevision + "/skills/engineering/"
	return []CatalogReference{
		{"setup-matt-pocock-skills", "somente referência", "Configura tracker, labels e instruções do projeto; exige decisão e autorização por repositório.", matt + "setup-matt-pocock-skills"},
		{"implement-spec", "somente referência", "Depende de spec, tickets, tracker e fluxo de PR do autor; não adaptada ao GEX.", matt + "implement-spec"},
		{"ask-matt", "somente referência", "Roteador pessoal presume catálogo e comandos completos do autor.", matt + "ask-matt"},
		{"impeccable", "por função", "Skill externa de UI envolve launcher, hooks e telemetria opt-out; não instalada.", "https://github.com/pbakaus/impeccable/tree/dad8dbb3aac21ad1fa236c110315ef4abece23e8"},
		{"security-audit", "por função", "Auditoria externa exige pedido explícito e sandbox isolado; não instalada.", "https://github.com/cloudflare/security-audit-skill/tree/c1c8a8c1471069fb0e188eeaff69b8e8db6564a8/skills/security-audit"},
	}
}

func HookCandidates() []CatalogReference {
	const claude = "https://github.com/davila7/claude-code-templates/tree/100b8bcbb2800b24cb33b6d5e5606c302fff198c/cli-tool/components/hooks/"
	const codex = "https://github.com/am-will/codex-skills/tree/9f954c3b63fda0d1875252009c196bf15072c1ca/hooks/aitmpl-codex/"
	const git = "https://git-scm.com/docs/githooks"
	candidates := []CatalogReference{
		{"gex-commit-msg", "por função", "Git hook por projeto; não é hook global de agente.", git + "#_commit_msg"},
		{"gex-pre-commit", "por função", "Git hook por projeto; requer regras e instalação próprias.", git + "#_pre_commit"},
		{"gex-pre-push", "por função", "Git hook por projeto; não substitui proteção remota.", git + "#_pre_push"},
		{"gex-session-start", "somente referência", "Instruções Core globais já carregadas; evento não requer hook redundante.", "https://code.claude.com/docs/en/hooks"},
		{"gex-sensitive-action-warning", "somente referência", "Conceito de alerta ainda sem adaptador seguro para Claude/Codex.", "https://learn.chatgpt.com/docs/hooks"},
		{"gex-pre-pr", "somente referência", "Pré-PR é fluxo de skill, não existe hook Git nativo de abertura de PR.", git},
	}
	claudeTemplates := []struct{ name, path, reason string }{
		{"lint-on-save", "development-tools/lint-on-save.json", "Pode reescrever arquivos e baixar ESLint; requer adaptação."},
		{"smart-formatting", "development-tools/smart-formatting.json", "Formata arquivos automaticamente e omite falhas."},
		{"change-tracker", "development-tools/change-tracker.json", "Registra caminhos no HOME sem retenção definida."},
		{"nextjs-code-quality-enforcer", "development-tools/nextjs-code-quality-enforcer.json", "Depende de jq e heurísticas específicas de Next.js."},
		{"run-tests-after-changes", "post-tool/run-tests-after-changes.json", "Executa script npm do projeto após edição; não cobre outras alterações."},
		{"security-scanner", "security/security-scanner.json", "Alertas não bloqueiam nem comprovam ausência de segredos."},
		{"smart-commit", "git-workflow/smart-commit.json", "Auto-commita edições; conflita com revisão e Conventional Commits GEX."},
		{"tdd-gate", "quality-gates/tdd-gate.json", "Presença de teste homônimo não prova cobertura; requer script e convenções."},
		{"scope-guard", "quality-gates/scope-guard.json", "Depende de spec recente; pode deixar alterações fora do escopo passarem."},
		{"env-file-protection", "security/env-file-protection.json", "Proteção pontual de Write não cobre outras operações com segredos."},
		{"format-javascript-files", "post-tool/format-javascript-files.json", "Reescreve arquivo e npx pode baixar dependência."},
		{"subagent-read-only-guard", "security/subagent-read-only-guard.json", "Depende de campo agent_id e script Python; não protege leitura de segredos."},
		{"change-logger", "automation/change-logger.json", "Registra comandos e caminhos potencialmente sensíveis no projeto."},
	}
	for _, template := range claudeTemplates {
		candidates = append(candidates, CatalogReference{"claude/" + template.name, "somente referência", template.reason + " Template não instalado nem validado em ambiente GEX.", claude + template.path})
	}
	codexBundles := []struct{ name, path, reason string }{
		{"agents-md-loader", "automation/agents-md-loader", "Pode duplicar contexto AGENTS.md já nativo."},
		{"build-on-change", "automation/build-on-change", "Executa build do projeto a cada edição e oculta falhas."},
		{"change-logger", "automation/change-logger", "Registra comandos potencialmente sensíveis no projeto."},
		{"simple-notifications", "automation/simple-notifications", "Depende de variável Claude não exportada pelo wrapper Codex."},
		{"change-tracker", "development-tools/change-tracker", "Registra caminhos no HOME sem retenção definida."},
		{"debug-window", "development-tools/debug-window", "Abre terminal e encerra processos; eventos não equivalentes."},
		{"lint-on-save", "development-tools/lint-on-save", "Pode reescrever arquivo e baixar ESLint."},
		{"validate-branch-name", "git/validate-branch-name", "Impõe Git Flow e não cobre outros comandos de branch."},
		{"console-log-cleaner", "pre-tool/console-log-cleaner", "Apenas avisa e pode ignorar arquivos do patch."},
		{"tdd-gate", "quality-gates/tdd-gate", "Filtra Edit/Write, mas Codex reporta apply_patch; pode ficar inerte."},
		{"secret-scanner", "security/secret-scanner", "Pode exibir trecho de segredo em stderr."},
		{"security-scanner", "security/security-scanner", "Falhas omitidas; ausência de aviso não prova segurança."},
	}
	for _, bundle := range codexBundles {
		candidates = append(candidates, CatalogReference{"codex/" + bundle.name, "somente referência", bundle.reason + " Bundle não instalado nem validado no Codex.", codex + bundle.path})
	}
	return candidates
}
