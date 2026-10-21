package app

import (
	"strings"
	"testing"
)

func credentialsClear(t *testing.T) {
	t.Helper()
	for _, name := range []string{"PAGAMERICAN_API_KEY", "BUYGOODS_ACCOUNT_ID", "BUYGOODS_API_TOKEN", "CLICKBANK_API_KEY", "CLICKBANK_ACCOUNT"} {
		t.Setenv(name, "")
	}
}

func TestCheckCredentialsNamesMissingVariableWithoutValue(t *testing.T) {
	credentialsClear(t)
	t.Setenv("BUYGOODS_ACCOUNT_ID", "conta-sigilosa")
	err := CheckCredentials([]string{"gex-buygoods-api"})
	if err == nil || !strings.Contains(err.Error(), "BUYGOODS_API_TOKEN") || strings.Contains(err.Error(), "conta-sigilosa") {
		t.Fatalf("variável faltante omitida ou valor exposto: %v", err)
	}
}

func TestCheckCredentialsRejectsEachGatewayWithoutEnvironment(t *testing.T) {
	for _, tc := range []struct{ skill, missing string }{
		{"gex-pagamerican-api", "PAGAMERICAN_API_KEY"},
		{"gex-buygoods-api", "BUYGOODS_ACCOUNT_ID"},
		{"gex-clickbank-api", "CLICKBANK_ACCOUNT"},
	} {
		credentialsClear(t)
		err := CheckCredentials([]string{tc.skill})
		if err == nil || !strings.Contains(err.Error(), tc.missing) || !strings.Contains(err.Error(), "canal corporativo") {
			t.Fatalf("%s sem credencial: %v", tc.skill, err)
		}
	}
}

func TestCheckCredentialsAcceptsPresentVariablesAndSkillsWithoutCredentials(t *testing.T) {
	credentialsClear(t)
	for _, name := range []string{"PAGAMERICAN_API_KEY", "BUYGOODS_ACCOUNT_ID", "BUYGOODS_API_TOKEN", "CLICKBANK_API_KEY", "CLICKBANK_ACCOUNT"} {
		t.Setenv(name, "presente")
	}
	if err := CheckCredentials([]string{"gex-buygoods-api", "gex-clickbank-api", "gex-local-spark-tests", "tdd"}); err != nil {
		t.Fatalf("presença basta para instalar: %v", err)
	}
}
