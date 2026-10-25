package catalog

import "slices"

type Specialty struct {
	Name        string
	Group       string
	Description string
	Warning     string
	Requires    []string
}

var specialtyList = []Specialty{
	{
		Name:        "gex-local-spark-tests",
		Group:       "Dados e Plataforma",
		Description: "Autoria GEX; fonte: workflow canonical-ci.yml do projeto gex-infra-aws. Roteiro local de testes PySpark/Iceberg.",
		Warning:     "Exige runtime PySpark/Iceberg local; runtime ausente é bloqueio, não sucesso.",
	},
	{
		Name:        "gex-canonical-prod-readiness",
		Group:       "Dados e Plataforma",
		Description: "Autoria GEX; fonte: skill interna de prontidão da Canonical (gex-infra-aws). Verificação somente leitura antes de regenerar Gold ou publicar no Hub.",
		Warning:     "Exige perfil AWS somente leitura de produção (Step Functions e Glue) obtido por canal aprovado; sem ele, a skill não conclui prontidão.",
	},
	{
		Name:        "gex-hub-gold-refresh-prod",
		Group:       "Dados e Plataforma",
		Description: "Autoria GEX; fonte: skill interna de atualização Gold e snapshot do Hub. Execução em produção só com autorização explícita; sem permissão, diagnóstico.",
		Warning:     "Executa state machines em produção: exige perfil AWS com permissão de execução e autorização explícita a cada uso; sem isso, entregue só diagnóstico.",
	},
	{
		Name:        "gex-gateway-audit",
		Group:       "Dados e Plataforma",
		Description: "Autoria GEX; fonte: método de auditoria profunda do vault canônico de gateways. Separa defeito de régua; registro e issues só com autorização.",
		Warning:     "Exige clone autorizado do vault canônico, perfil AWS somente leitura de validação e, para `--api`, credenciais de gateway do próprio vault; a skill não traz ferramenta nem credencial.",
	},
	{
		Name:        "gex-pr-review",
		Group:       "Dados e Plataforma",
		Description: "Autoria GEX; fonte: skill interna de revisão de PR do gex-infra-aws. Complementa gex-revisar-pr com SHA fixo, norma CAN/CHK e prova PySpark/Iceberg local.",
		Warning:     "Exige `gh` autenticado com leitura no gex-infra-aws e no vault canônico; publicar review ou mudar labels exige escrita e autorização explícita.",
		Requires:    []string{"gex-local-spark-tests"},
	},
	{
		Name:        "gex-design-system",
		Group:       "Produto e Interface",
		Description: "Autoria GEX; fonte: Design System GEX exportado (tokens.css e logos SVG). Identidade visual para relatórios, páginas e slides fora do Hub.",
		Warning:     "Sem credencial. Marca GEX para materiais internos; licença formal de uso da marca não confirmada: consulte o responsável pela marca.",
	},
	{
		Name:        "gex-hub-design-system",
		Group:       "Produto e Interface",
		Description: "Autoria GEX; fonte: convenções do gex-hub-dashboards (tokens locais, shadcn/ui, Storybook). UI dentro do Hub.",
		Warning:     "Sem credencial. `shadcn add` baixa código de terceiros: rode `--dry-run` e revise antes.",
	},
}

func Specialties() []Specialty {
	out := make([]Specialty, len(specialtyList))
	for i, specialty := range specialtyList {
		specialty.Requires = slices.Clone(specialty.Requires)
		out[i] = specialty
	}
	return out
}

func SpecialtyGroups() []string {
	return []string{"Dados e Plataforma", "Produto e Interface"}
}

func SpecialtyByName(name string) (Specialty, bool) {
	for _, specialty := range Specialties() {
		if specialty.Name == name {
			return specialty, true
		}
	}
	return Specialty{}, false
}
