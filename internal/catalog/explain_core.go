package catalog

func init() {
	registerExplanations(map[string]Explanation{
		"gex-abrir-issue": {
			What:    "Transforma demanda ou bug em issue com problema, evidência e critérios de aceite; verifica duplicatas, assignee e labels antes de publicar com autorização.",
			UseCase: "Você encontrou um erro no cálculo de vendas e precisa registrar reprodução e resultado esperado para outra pessoa corrigir sem adivinhar a regra.",
			Analogy: "É como abrir uma ordem de serviço: descreve o defeito e como conferir o conserto, sem inventar a peça que deve ser trocada.",
		},
		"gex-auto-revisao": {
			What:    "Compara seu próprio diff com a issue, critérios de aceite, testes e riscos antes da PR; entrega achados fundamentados sem editar nem publicar.",
			UseCase: "Você terminou uma alteração no pipeline e quer identificar um caminho de erro esquecido antes de pedir revisão para a equipe.",
			Analogy: "É a conferência da própria mala antes de fechar: verificar a lista e o que entrou, sem pedir que outra pessoa faça isso primeiro.",
		},
		"gex-fluxo-desenvolvimento": {
			What:    "Organiza uma mudança GEX da demanda à issue, implementação, verificação, auto-revisão e PR, ajustando os passos ao risco e às regras do projeto.",
			UseCase: "Você começou uma correção em um job de dados e precisa decidir quais critérios, testes e revisões antecedem a entrega da PR.",
			Analogy: "É um roteiro de viagem que indica paradas necessárias conforme o trajeto; não substitui as regras de trânsito de cada lugar.",
		},
		"gex-preparar-pr": {
			What:    "Confere diff, issue real, auto-revisão, testes, assignee e reviewers; prepara título e descrição com evidências e riscos antes da publicação autorizada.",
			UseCase: "Você vai abrir PR de uma mudança no adaptador de gateway e quer garantir vínculo correto com a issue e provas de teste executadas.",
			Analogy: "É o checklist antes de despachar uma encomenda: confere conteúdo, destino e documentos antes de entregá-la ao transporte.",
		},
		"gex-regras-de-negocio": {
			What:    "Rastreia regras confirmadas da fonte até transformações e testes; explicita entradas, saídas, duplicidades e lacunas sem aprovar hipóteses como regra.",
			UseCase: "Você altera um join da Silver e precisa provar que pedidos duplicados não multiplicam receita nem distorcem totais de vendas.",
			Analogy: "É comparar receita culinária, preparo e prato servido: se a medida original falta, não se inventa a quantidade do ingrediente.",
		},
		"gex-revisar-pr": {
			What:    "Revisa PR publicada por SHA contra issue, diff, contratos e testes; aponta defeitos concretos nas linhas alteradas quando a publicação está autorizada.",
			UseCase: "Você recebeu PR que muda tratamento de refunds e precisa localizar regressão reproduzível antes de comentar no GitHub.",
			Analogy: "É vistoria de obra feita sobre uma planta e uma versão datada: cada defeito aponta o lugar exato, sem repetir reclamações genéricas.",
		},
		"gex-verificar-mudanca": {
			What:    "Liga critérios de aceite a comportamentos e verificações executadas, checa erros e registra resultados ou bloqueios sem chamar leitura de código de teste.",
			UseCase: "Você concluiu ajuste no Glue e precisa registrar quais testes passaram e qual cenário de falha não pôde verificar antes de anunciar entrega.",
			Analogy: "É inspeção de entrega com comprovantes: cada item da lista precisa de uma checagem observável, não apenas de uma assinatura.",
		},
		"gex-local-spark-tests": {
			What:    "Orienta preparar PySpark e Iceberg em ambiente local isolado e rodar testes reais do checkout, sem consultar serviços AWS nem simular job remoto.",
			UseCase: "Você alterou escrita Iceberg do canônico e precisa executar replay e checar isolamento de outro gateway antes da revisão da PR.",
			Analogy: "É testar uma máquina numa bancada separada com peças reais do projeto, sem ligá-la à linha de produção.",
		},
		"gex-canonical-prod-readiness": {
			What:    "Consulta em modo somente leitura execuções Canonical, jobs Glue, marcadores Gold e publicação Hub para decidir se a Silver está pronta; não dispara jobs.",
			UseCase: "Você precisa liberar atualização do dashboard após reprocessamento e quer confirmar que Canonical terminou antes da Gold e da extração do Hub.",
			Analogy: "É vistoriar cada estação de uma linha de montagem antes de liberar a entrega; uma estação verde isolada não aprova a linha inteira.",
		},
		"gex-hub-gold-refresh-prod": {
			What:    "Após autorização explícita, executa atualização da Gold e extração do Hub em ordem, checa marcadores e confirma ingestão; sem permissão, entrega diagnóstico.",
			UseCase: "Você precisa atualizar snapshot do Hub após mudança Canonical aprovada e conferir se o worker publicou dados novos, não só enfileirou mensagem.",
			Analogy: "É repor uma vitrine: primeiro preparar o produto, depois transportar e conferir que chegou à prateleira, não só ao caminhão.",
		},
		"gex-gateway-audit": {
			What:    "Usa ferramentas do vault autorizado para inventariar cargas, rodar auditoria profunda e separar defeitos de falsos FAILs, guardando contraprovas e retro.",
			UseCase: "Você investiga FAILs no gateway CartPanda e precisa distinguir erro de cálculo real de uma régua com coluna inexistente antes de propor issue.",
			Analogy: "É calibrar uma balança antes de acusar falta no estoque: medida errada e produto faltando pedem respostas diferentes.",
		},
		"gex-pr-review": {
			What:    "Revisa PR do gex-infra-aws por SHA e norma CAN/CHK, executa o caminho alterado em PySpark/Iceberg local e classifica achados com evidência.",
			UseCase: "Você revisa uma PR de fulfillment e precisa provar que o SQL real do job aceita replay sem duplicar dados antes de emitir parecer.",
			Analogy: "É testar a peça substituída numa bancada com o manual vigente, anotando número da versão antes de assinar a inspeção.",
		},
		"gex-design-system": {
			What:    "Fornece tokens CSS e logos SVG da identidade GEX para criar páginas, relatórios e slides nos temas claro e escuro fora do código do Hub.",
			UseCase: "Você cria um relatório visual interno sobre receita por gateway e precisa manter cores, tipografia e contraste consistentes com a marca GEX.",
			Analogy: "É um mostruário de tintas e medidas aprovado para montar peças diferentes com a mesma identidade, sem escolher cores de improviso.",
		},
		"gex-hub-design-system": {
			What:    "Orienta componentes e páginas no repositório do Hub com tokens locais, variantes shadcn/ui e stories Storybook, preservando API e acessibilidade.",
			UseCase: "Você adiciona um card ao dashboard do Hub e precisa usar tokens atuais e documentar variantes no Storybook sem criar cores soltas.",
			Analogy: "É usar peças e medidas da oficina onde a tela será montada, em vez de trazer um catálogo antigo de outra oficina.",
		},
		"gex-conventional-commits": {
			What:    "Antes de comando git commit por agente em repositório GEX, analisa a mensagem e bloqueia formatos fora de Conventional Commits ou comandos impossíveis de analisar.",
			UseCase: "Você pede ao agente commit de correção no pipeline e quer impedir mensagem sem tipo e descrição antes de executar o comando.",
			Analogy: "É porteiro que confere etiqueta na caixa antes de ela sair pela porta; não substitui conferência do conteúdo depois.",
		},
		"gex-sensitive-action-warning": {
			What:    "No Claude Code em repositório GEX, pede confirmação antes de comandos sensíveis reconhecidos, como exclusão, publicação externa ou leitura de segredo.",
			UseCase: "Você pede análise de arquivos e o agente tenta ler um .env ou rodar um comando destrutivo: recebe pergunta antes da ação.",
			Analogy: "É aviso na porta de uma área restrita: pede sua confirmação antes de entrar, sem executar a tarefa por você.",
		},
		"gex-pre-pr": {
			What:    "Antes de gh pr create pelo agente em repositório GEX, pede confirmação e lembra auto-revisão, testes, build, lint, issue, assignee e reviewers.",
			UseCase: "Você está prestes a abrir PR do adaptador e quer ser lembrado de conferir evidências e responsáveis antes da publicação.",
			Analogy: "É a pergunta do balcão antes de protocolar documento: confere lista de anexos, mas não revisa o texto por conta própria.",
		},
		"gex-session-start": {
			What:    "No começo da sessão, informa se o diretório é repositório GEX, quantas skills Core estão instaladas e se o bloco de instruções Core existe; não altera arquivos.",
			UseCase: "Você abre sessão de agente em um checkout novo e quer saber se o kit e as instruções Core estão disponíveis antes de começar.",
			Analogy: "É quadro de avisos na entrada da oficina: mostra o ambiente disponível, sem instalar ferramentas nem iniciar trabalho.",
		},
		"gex-commit-msg": {
			What:    "Quando o Git executa commit-msg, valida a mensagem recebida contra Conventional Commits; se o binário do kit estiver ausente, avisa e não bloqueia.",
			UseCase: "Você faz git commit manual no projeto e quer detectar mensagem fora do padrão no momento em que o Git recebe o texto.",
			Analogy: "É carimbo na saída do correio que confere formato do endereço; sem o funcionário presente, só deixa aviso.",
		},
		"gex-pre-commit": {
			What:    "Antes do commit pelo Git, procura possível credencial no diff e roda comandos rápidos documentados configurados; binário ausente gera aviso e pula o hook.",
			UseCase: "Você vai commitar uma mudança no ETL e quer barrar um token acidental no diff e rodar verificações rápidas do projeto.",
			Analogy: "É inspeção de bagagem na porta: confere item proibido e uma lista de checagens antes de liberar a saída.",
		},
		"gex-pre-push": {
			What:    "Antes do git push, executa os comandos documentados pelo projeto que foram configurados para esse hook; falha de comando interrompe o push.",
			UseCase: "Você vai enviar branch de correção e quer executar verificações do projeto escolhidas antes de transferir os commits ao remoto.",
			Analogy: "É a inspeção antes de carregar um caminhão: executa as checagens combinadas antes de deixar o veículo partir.",
		},
	})
}
