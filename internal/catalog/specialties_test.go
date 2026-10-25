package catalog

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var specialtiesPersonalData = regexp.MustCompile(`/Users/|/home/|~/vaults|davi-agents|Davi|Oliveira|Selinger|gex-terraform|canonico-validador|TickTick|skill://|agent://|@latest|Male Enhancement|<metadata>|C2PA|potrace|vaul-canonico-gateways|canonico-gateways|Cockpit|arn:aws:|--profile|AWS_PROFILE|ghp_[A-Za-z0-9]{20,}|github_pat_|AKIA[0-9A-Z]{16}|[0-9]{12}`)

var specialtiesEmail = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+[.][A-Za-z]{2,}`)

var specialtiesReservedDomain = regexp.MustCompile(`@example[.](com|org|net)$`)

func specialtiesLeak(data []byte) []byte {
	if match := specialtiesPersonalData.Find(data); match != nil {
		return match
	}
	for _, match := range specialtiesEmail.FindAll(data, -1) {
		if !specialtiesReservedDomain.Match(match) {
			return match
		}
	}
	return nil
}

func TestSpecialtyAssetsHaveNoPersonalData(t *testing.T) {
	files := 0
	err := filepath.WalkDir("assets", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		files++
		if match := specialtiesLeak(data); match != nil {
			t.Fatalf("%s contém dado pessoal, caminho local ou token: %q", name, match)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("não foi possível ler os assets: %v", err)
	}
	if files == 0 {
		t.Fatalf("nenhum asset foi lido")
	}
}

func TestSpecialtyPersonalDataPatternCatchesLeaks(t *testing.T) {
	leaks := []string{"/home/usuario/projeto", "usuario@exemplo.com.br", "123456789012", "arn:aws:iam::123456789012:role/x"}
	for _, leak := range leaks {
		if specialtiesLeak([]byte(leak)) == nil {
			t.Fatalf("padrão deixou passar: %s", leak)
		}
	}
	if specialtiesLeak([]byte("private@example.com")) != nil {
		t.Fatalf("domínio reservado de exemplo bloqueado")
	}
	if specialtiesLeak([]byte("Consulte a skill aplicável quando disponível.")) != nil {
		t.Fatalf("padrão bloqueou texto comum")
	}
}

func TestSpecialtyCatalogGroupsAndDescriptions(t *testing.T) {
	specialties := Specialties()
	if len(specialties) != 7 || len(SpecialtySkills()) != 7 {
		t.Fatalf("especializações inesperadas: %d", len(specialties))
	}
	counts := map[string]int{}
	for _, specialty := range specialties {
		if !slices.Contains(SpecialtyGroups(), specialty.Group) {
			t.Fatalf("grupo inválido em %s: %s", specialty.Name, specialty.Group)
		}
		if !strings.HasPrefix(specialty.Description, "Autoria GEX; fonte: ") {
			t.Fatalf("descrição fora do padrão: %s", specialty.Name)
		}
		if specialty.Warning == "" {
			t.Fatalf("especialização sem aviso: %s", specialty.Name)
		}
		counts[specialty.Group]++
	}
	want := map[string]int{"Dados e Plataforma": 5, "Produto e Interface": 2}
	for group, count := range want {
		if counts[group] != count {
			t.Fatalf("grupo %s com %d especializações, esperado %d", group, counts[group], count)
		}
	}
	if _, ok := SpecialtyByName("inexistente"); ok {
		t.Fatal("especialização inexistente encontrada")
	}
}

func TestPrReviewRequiresLocalSparkTests(t *testing.T) {
	specialty, ok := SpecialtyByName("gex-pr-review")
	if !ok || !slices.Equal(specialty.Requires, []string{"gex-local-spark-tests"}) {
		t.Fatalf("gex-pr-review sem dependência declarada: %+v", specialty)
	}
	if err := ValidateSelection([]string{"gex-pr-review"}, nil); err == nil || !strings.Contains(err.Error(), "gex-local-spark-tests") {
		t.Fatalf("gex-pr-review aceita sem gex-local-spark-tests: %v", err)
	}
	if err := ValidateSelection([]string{"gex-pr-review", "gex-local-spark-tests"}, nil); err != nil {
		t.Fatalf("par obrigatório rejeitado: %v", err)
	}
	if !strings.Contains(SkillDescription("gex-pr-review"), "Requer seleção conjunta: gex-local-spark-tests.") {
		t.Fatalf("descrição sem aviso de seleção conjunta: %s", SkillDescription("gex-pr-review"))
	}
}
