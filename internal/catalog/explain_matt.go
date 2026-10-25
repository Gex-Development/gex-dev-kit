package catalog

func init() {
	registerExplanations(map[string]Explanation{
		"codebase-design": {
			What:    "Ajuda a desenhar módulos que escondem bastante comportamento atrás de uma interface pequena. Examina o que cada chamada precisa saber, onde fica a divisão entre módulos e como testar pelo comportamento visível.",
			UseCase: "Ao redesenhar um módulo de pedidos com muitos métodos e regras espalhadas, você compara interfaces para concentrar as regras sem complicar quem chama.",
			Analogy: "Como um balcão de atendimento: você faz um pedido simples, e o trabalho de várias etapas acontece atrás do balcão sem exigir que conheça cada funcionário.",
		},
		"code-review": {
			What:    "Compara as mudanças desde um commit ou ramo escolhido com dois critérios separados: as regras de código do repositório e o pedido original. Reúne os achados das duas revisões lado a lado.",
			UseCase: "Antes de integrar uma PR, você informa o ponto de comparação e a issue; a revisão aponta tanto padrões quebrados quanto requisitos que ficaram de fora.",
			Analogy: "Como vistoriar uma reforma com duas listas: uma confere normas da obra; a outra, se foi entregue exatamente o que o morador pediu.",
		},
		"diagnosing-bugs": {
			What:    "Monta primeiro um teste ou comando que reproduz o defeito de forma confiável. Depois reduz o caso, testa hipóteses com medidas ou inspeções e confirma que a correção elimina o sintoma.",
			UseCase: "Uma exportação falha só em alguns registros: você cria um caso que falha igual, isola o registro responsável e usa esse sinal para testar a correção.",
			Analogy: "Como mecânico que reproduz o barulho na oficina antes de trocar peças: cada teste elimina uma causa possível até o ruído desaparecer.",
		},
		"domain-modeling": {
			What:    "Esclarece termos ambíguos do negócio com exemplos concretos e confronta o vocabulário com o código. Registra termos resolvidos no glossário e decisões difíceis de reverter em registros de decisão quando cabível.",
			UseCase: "Ao discutir cancelamento de pedidos, você distingue cancelar um item de cancelar o pedido inteiro e registra os significados antes de alterar o modelo.",
			Analogy: "Como montar um dicionário comum antes de redigir um contrato: todos passam a usar a mesma palavra para o mesmo fato, sem negociar o sentido a cada conversa.",
		},
		"grill-me": {
			What:    "Aciona uma entrevista insistente sobre um plano ou decisão por meio da skill grilling. As perguntas percorrem as decisões possíveis em rodadas, trazendo uma resposta recomendada para cada pergunta.",
			UseCase: "Antes de escolher a arquitetura de uma fila, você expõe a ideia e responde às rodadas de perguntas sobre escolhas e consequências ainda em aberto.",
			Analogy: "Como pedir a um colega exigente que questione seu plano antes da reunião: ele não aceita um desenho com decisões importantes deixadas em branco.",
		},
		"grill-with-docs": {
			What:    "Combina a entrevista de grilling com domain-modeling: questiona decisões de um plano e registra, conforme surgem, termos esclarecidos no glossário e decisões que merecem registro.",
			UseCase: "Ao definir como pedidos e cobranças se relacionam, você responde perguntas sobre casos-limite e deixa a linguagem acordada documentada durante a conversa.",
			Analogy: "Como reunião de projeto em que alguém faz as perguntas difíceis e outro atualiza o mapa de termos e as decisões assim que o grupo chega a um acordo.",
		},
		"grilling": {
			What:    "Organiza uma ideia como uma árvore de decisões e entrevista você em rodadas. Pergunta primeiro o que já pode ser decidido, recomenda respostas e só avança para questões dependentes após receber suas escolhas.",
			UseCase: "Você propõe sincronizar dados entre sistemas; a entrevista separa decisões sobre origem dos dados, conflitos e falhas, sem presumir respostas para etapas seguintes.",
			Analogy: "Como planejar uma viagem com alguém criterioso: primeiro decide destino e datas; só então discute transporte e hospedagem, sempre mostrando as opções em aberto.",
		},
		"handoff": {
			What:    "Resume a conversa atual em um documento temporário para outro agente continuar. Aponta para planos e decisões já existentes em vez de copiá-los, sugere skills úteis e retira informações sensíveis.",
			UseCase: "Ao trocar de sessão no meio de uma implementação, você gera um resumo com o que foi decidido e onde estão os artefatos para o próximo agente retomar.",
			Analogy: "Como passagem de plantão: o próximo responsável recebe o estado do trabalho e os endereços dos registros, sem precisar reler toda a conversa.",
		},
		"improve-codebase-architecture": {
			What:    "Examina áreas recentes ou indicadas do código em busca de módulos cuja interface expõe complexidade demais. Mostra oportunidades de simplificação num relatório HTML visual e entrevista você sobre a opção escolhida.",
			UseCase: "Após várias mudanças espalhadas no fluxo de pedidos, você vê num relatório quais módulos repassam trabalho sem concentrar regras e escolhe um para discutir.",
			Analogy: "Como planta de uma casa antes da reforma: mostra onde há corredores demais para chegar a um cômodo e permite escolher qual trecho repensar primeiro.",
		},
		"pr": {
			What:    "Redige o corpo de uma PR com um resumo visual da mudança, evidências de antes e depois e os riscos de integrá-la. Usa o vocabulário do projeto e evita introduções longas.",
			UseCase: "Depois de corrigir um cálculo, você apresenta a diferença no fluxo, o teste que falhava e agora passa, e o impacto possível ao integrar a alteração.",
			Analogy: "Como etiqueta de entrega de uma reforma: mostra o que mudou, a prova de que funciona e o que pode dar errado ao abrir a porta ao público.",
		},
		"prototype": {
			What:    "Constrói um experimento descartável para responder uma dúvida de lógica ou aparência. Para regras, cria uma página interativa que expõe mudanças de estado; para interface, mostra variantes visuais alternáveis.",
			UseCase: "Antes de fixar o fluxo de aprovação, você testa estados e casos estranhos num HTML clicável; se a dúvida for visual, compara layouts no navegador.",
			Analogy: "Como maquete: você experimenta o funcionamento ou a disposição das peças antes de construir, aproveita a decisão validada e descarta a montagem de teste.",
		},
		"research": {
			What:    "Envia uma investigação para um agente em segundo plano, que consulta fontes primárias como documentação oficial e código. Registra as conclusões em um arquivo Markdown com uma fonte para cada afirmação.",
			UseCase: "Antes de usar uma API pouco conhecida, você pede pesquisa sobre seus limites; recebe um arquivo no repositório citando a documentação e os trechos relevantes.",
			Analogy: "Como encarregar alguém de consultar os manuais originais e entregar um caderno de notas com a página de origem ao lado de cada conclusão.",
		},
		"retro": {
			What:    "Lê os registros da sessão de programação indicada e sugere ajustes no ambiente do agente para as próximas execuções. Procura obstáculos de navegação, checagens ausentes e instruções que não ajudam.",
			UseCase: "Depois de uma sessão em que o agente demorou para achar arquivos e repetiu um erro, você recebe sugestões de ponteiros e checagens para evitar nova ocorrência.",
			Analogy: "Como reunião curta depois de um atendimento: em vez de refazer o trabalho, a equipe identifica o que mudar no posto para o próximo atendimento ser melhor.",
		},
		"tdd": {
			What:    "Orienta ciclos curtos de teste que falha e implementação que o faz passar. Os testes descrevem comportamento observado pela interface pública, evitando depender de detalhes internos ou repetir a própria lógica testada.",
			UseCase: "Ao acrescentar desconto no checkout, você escreve um teste com o total esperado, vê a falha e só então implementa a regra necessária para aquele caso.",
			Analogy: "Como definir o resultado esperado de uma receita antes de cozinhar: cada tentativa é conferida pelo prato pronto, não pelo número de vezes que mexeu a panela.",
		},
		"teach": {
			What:    "Organiza o diretório atual como espaço de estudo ligado ao motivo de aprender. Produz lições HTML curtas com prática interativa, fontes confiáveis e registros do que você demonstrou ter aprendido.",
			UseCase: "Você quer aprender SQL para investigar vendas; as lições usam consultas pequenas, exercícios com retorno imediato e referências para revisar depois.",
			Analogy: "Como professor particular com caderno de acompanhamento: ajusta o próximo exercício ao seu objetivo e ao que você já consegue fazer sozinho.",
		},
		"to-questionnaire": {
			What:    "Pergunta a quem será enviado um questionário e quais respostas você precisa obter. Cria um arquivo Markdown com perguntas direcionadas ao conhecimento dessa pessoa, prontas para resposta assíncrona ou reunião.",
			UseCase: "Uma integração depende das regras de cancelamento conhecidas pelo time financeiro; você gera perguntas objetivas para colher essas regras antes de programar.",
			Analogy: "Como preparar um formulário para o especialista certo: você define o que precisa decidir, e ele preenche apenas as lacunas que só ele pode esclarecer.",
		},
		"wait-what": {
			What:    "Pede ao agente que reapresente sua última explicação com um pouco mais de contexto, em inglês técnico simplificado e usando os termos definidos no glossário do projeto.",
			UseCase: "Uma explicação sobre a arquitetura ficou difícil de seguir; você chama a skill para receber a mesma ideia com contexto e frases mais diretas.",
			Analogy: "Como interromper uma apresentação e pedir: volte um passo, use palavras que combinamos e explique de novo sem pular a parte importante.",
		},
		"writing-for-agents": {
			What:    "Orienta a escrita de documentos que agentes consultam, como skills e instruções do repositório. Mostra como indicar quando ler cada documento, ordenar passos e referências e definir critérios claros de conclusão.",
			UseCase: "Ao escrever uma skill para revisar consultas SQL, você ajusta o texto que dispara sua leitura e deixa claros os passos e o sinal de que a revisão terminou.",
			Analogy: "Como manual de trabalho com índice útil: o índice diz quando abrir cada capítulo, e cada procedimento diz exatamente quando a tarefa acabou.",
		},
	})
}
