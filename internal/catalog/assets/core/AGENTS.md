# Instruções globais de desenvolvimento GEX

Use estas instruções como base global do trabalho GEX. Fora dele, não imponha convenções da empresa. Leia também as instruções do projeto e as orientações da tarefa antes de agir.

## Precedência e escopo

- Siga políticas da plataforma e instruções explícitas da tarefa. Instruções locais do repositório especializam stack, comandos, domínio e fluxo; não as apague nem substitua por este texto global. Se houver conflito real, pare, explique qual regra conflita e peça decisão ao responsável; não silencie o conflito.
- Não presuma acesso, autorização de publicação nem ferramentas instaladas. Use apenas integrações aprovadas para a pessoa e o projeto. Addons são opt-in: Caveman, Ponytail, Superpowers e SDD/Spec Kit não integram o Core nem se ativam por esta base. GSD não integra o kit.
- Skills de trabalho de autoria GEX usam prefixo `gex-`. Consulte a skill aplicável quando disponível, sem tratar este texto como prova de instalação.

## Escrita

- Comece pela conclusão ou ação necessária. Depois apresente contexto e evidência.
- Escreva frases curtas, use verbos concretos e divida assuntos em tópicos pequenos. Explique siglas quando o leitor precisar.
- Evite densidade, jargão gratuito e repetição. Adapte a explicação ao leitor sem perder rigor técnico.

## Trabalho e evidência

- Entenda a demanda e o comportamento atual antes de alterar código. Vincule mudança a critérios de aceite e explicite lacunas de regra de negócio.
- Preserve validação, tratamento de erro, segurança e integridade de dados. Mudanças em contratos, migrações e operações destrutivas exigem atenção aos consumidores e a plano de reversão. Não exponha credenciais, tokens, dados pessoais nem saídas sensíveis em código, logs, issues ou PRs; não peça segredos no chat. Antes de publicar ou executar ação irreversível, obtenha autorização conforme o projeto.
- Escreva ou atualize testes proporcionais ao risco e execute as verificações relevantes quando houver ambiente e autorização. Para cada critério, registre evidência observável: comando e resultado, reprodução, ou motivo explícito para não executar. Nunca afirme que um teste passou sem executá-lo nem transforme suposição em fato.
- No PR, registre escopo, issue vinculada, impacto, riscos, evidência e pendências. Antes de abrir, faça a auto-revisão e o preflight com testes, build e lint aplicáveis definidos pelo projeto. Registre comando e resultado real; falha ou ausência de comando/aplicabilidade exige explicação, não sucesso presumido. Não publique antes da correção ou exceção explícita decidida pelo responsável.

## Código e autoria

- Use Conventional Commits para commits de projetos GEX: `tipo(escopo opcional): descrição`, com `feat` para nova funcionalidade e `fix` para correção; outros tipos, como `docs`, `refactor`, `test`, `chore`, `perf`, `build`, `ci` ou `revert`, são permitidos. Use `!` ou rodapé `BREAKING CHANGE:` para quebra de contrato. Mensagem precisa descrever mudança real, não prometer validação que não ocorreu.
- Em código novo ou alterado, não escreva comentários nem docstrings explicativas, inclusive em testes, scripts, SQL e configuração. Explique decisões em documentação ou PR. Preserve diretivas técnicas necessárias, avisos legais e comentários legados que não estejam sendo alterados; não os remova apenas para cumprir esta regra.
- Não acrescente rótulos de geração, menção à ferramenta de IA usada nem coautoria automática de IA a código, testes, documentação da mudança, commits ou PRs. Identifique o responsável real. Preserve divulgação exigida por norma ou contrato e referências técnicas necessárias ao produto. Nunca declare falsamente que não houve uso de IA.

## Limites dos hooks

Hooks Git locais podem ser contornados, e Git não possui hook nativo antes de abrir PR. A revisão anterior à PR acontece no fluxo CLI/agente, não no pre-push. Hooks locais não substituem proteção remota.
