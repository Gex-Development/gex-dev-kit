package hooks

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

func precommitDiff(path string, start int, lines ...string) string {
	var b strings.Builder
	b.WriteString("diff --git a/" + path + " b/" + path + "\n")
	b.WriteString("index 1111111..2222222 100644\n")
	b.WriteString("--- /dev/null\n")
	b.WriteString("+++ b/" + path + "\n")
	b.WriteString("@@ -0,0 +" + strconv.Itoa(start) + "," + strconv.Itoa(len(lines)) + " @@\n")
	for _, line := range lines {
		b.WriteString("+" + line + "\n")
	}
	return b.String()
}

func TestScanStagedEncontraCredenciaisComArquivoELinha(t *testing.T) {
	diff := precommitDiff("config/app.env", 10,
		"nome = 'valor'",
		"aws = 'AKI"+"A"+"ABCDEFGHIJKLMNOP'",
		"token = 'gh"+"p_"+"abcdefghijklmnopqrstuvwxyz0123456789'",
		"-----BEGIN "+"RSA PRIVATE KEY-----",
		"slack = 'xox"+"b-"+"1234567890-abcdef'",
		"openai = 'sk"+"-"+"abcdefghijklmnopqrstuvwxyzABCDEF'",
		"google = 'AI"+"za"+"0123456789abcdefghijklmnopqrstuvwxy'",
		"<<<<<<< "+"HEAD",
	)
	findings, truncated, err := ScanStaged(strings.NewReader(diff), 1<<20)
	if err != nil {
		t.Fatalf("ScanStaged devolveu erro: %v", err)
	}
	if truncated {
		t.Fatalf("diff pequeno não deveria estar truncado")
	}
	want := []Finding{
		{Path: "config/app.env", Line: 11, Rule: "aws-access-key"},
		{Path: "config/app.env", Line: 12, Rule: "github-token"},
		{Path: "config/app.env", Line: 13, Rule: "private-key"},
		{Path: "config/app.env", Line: 14, Rule: "slack-token"},
		{Path: "config/app.env", Line: 15, Rule: "openai-key"},
		{Path: "config/app.env", Line: 16, Rule: "google-api-key"},
		{Path: "config/app.env", Line: 17, Rule: "conflict-marker"},
	}
	if !slices.Equal(findings, want) {
		t.Fatalf("achados = %+v; esperado %+v", findings, want)
	}
}

func TestScanStagedIgnoraBinarioELinhaAcimaDoLimite(t *testing.T) {
	longa := strings.Repeat("a", precommitMaxLine) + "AKI" + "A" + "ABCDEFGHIJKLMNOP"
	diff := "diff --git a/img.png b/img.png\nBinary files /dev/null and b/img.png differ\n" + precommitDiff("grande.txt", 1, longa)
	findings, _, err := ScanStaged(strings.NewReader(diff), 1<<22)
	if err != nil {
		t.Fatalf("ScanStaged devolveu erro: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("binário e linha acima do limite deveriam ser ignorados; achados = %+v", findings)
	}
}

func TestScanStagedMarcaTruncadoAoEstourarLimite(t *testing.T) {
	diff := precommitDiff("a.txt", 1, "linha qualquer", "outra linha")
	_, truncated, err := ScanStaged(strings.NewReader(diff), 10)
	if err != nil {
		t.Fatalf("ScanStaged devolveu erro: %v", err)
	}
	if !truncated {
		t.Fatalf("diff maior que o limite deveria ser marcado como truncado")
	}
}

func TestScanStagedRecusaDiffAusente(t *testing.T) {
	if _, _, err := ScanStaged(nil, 10); err == nil {
		t.Fatalf("diff ausente deveria devolver erro")
	}
}

func TestScanStagedMensagemCitaArquivoRegraEExcecaoSemValor(t *testing.T) {
	finding := Finding{Path: "config/app.env", Line: 11, Rule: "aws-access-key"}
	msg := finding.Message("~/repo/.git/gex-dev-kit/pre-commit-allow")
	for _, want := range []string{"config/app.env:11", "aws-access-key", "~/repo/.git/gex-dev-kit/pre-commit-allow"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("mensagem sem %q: %s", want, msg)
		}
	}
	if strings.Contains(msg, "AKIA") {
		t.Fatalf("mensagem não pode repetir o valor da credencial: %s", msg)
	}
}
