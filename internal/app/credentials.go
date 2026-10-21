package app

import (
	"fmt"
	"os"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

func CheckCredentials(skills []string) error {
	for _, skill := range skills {
		specialty, ok := catalog.SpecialtyByName(skill)
		if !ok {
			continue
		}
		for _, name := range specialty.Credentials {
			if os.Getenv(name) == "" {
				return fmt.Errorf("%s exige %s no ambiente. Solicite acesso pelo canal corporativo aprovado; nenhum arquivo foi aplicado.", skill, name)
			}
		}
	}
	return nil
}
