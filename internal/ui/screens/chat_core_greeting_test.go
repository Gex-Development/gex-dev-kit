package screens

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

func TestCoreGreetingNeverDeniesAnExistingCore(t *testing.T) {
	for _, tc := range []struct {
		name string
		core app.CoreStatus
		want string
	}{
		{"sem Core", app.CoreStatus{State: "não instalado"}, "Ainda não há Core"},
		{"Core em dia", app.CoreStatus{ActiveCore: 7, Checked: 20, State: "íntegro"}, "Encontrei o Core instalado (7/7"},
		{"Core alterado sem skill em dia", app.CoreStatus{Attention: 3, Checked: 2, State: "alterado"}, "3 arquivo(s) que precisam de atenção (0/7 skills em dia). Rode /doctor"},
		{"Core parcial com alteração", app.CoreStatus{ActiveCore: 4, Attention: 1, State: "alterado"}, "1 arquivo(s) que precisam de atenção (4/7"},
	} {
		got := coreGreeting(tc.core)
		if !strings.Contains(got, tc.want) {
			t.Fatalf("%s: %q não contém %q", tc.name, got, tc.want)
		}
		if tc.core.Attention > 0 && strings.Contains(got, "Ainda não há") {
			t.Fatalf("%s: negou um Core que existe: %q", tc.name, got)
		}
	}
}
