package agents

import (
	"strings"
	"testing"
)

func TestReadLatestBodyFitsTheRealCodexMetadataButStaysTightForClaude(t *testing.T) {
	asset := `{"name":"codex-aarch64-apple-darwin.tar.gz","digest":"sha256:0000000000000000000000000000000000000000000000000000000000000000","browser_download_url":"https://releases.openai.com/codex/releases/0.164.0/codex-aarch64-apple-darwin.tar.gz"},`
	codex := `{"assets":[` + strings.Repeat(asset, 400) + `{"name":"last"}],"tag_name":"rust-v0.164.0"}`
	if len(codex) < 90<<10 {
		t.Fatalf("o corpo de teste precisa se parecer com os ~97 KB reais, tem %d bytes", len(codex))
	}
	body, err := readLatestBody(strings.NewReader(codex), "codex")
	if err != nil {
		t.Fatalf("metadado real do Codex recusado: %v", err)
	}
	if got, err := parseLatest("codex", body); err != nil || got != "0.164.0" {
		t.Fatalf("versão do Codex: %q, %v", got, err)
	}
	if _, err := readLatestBody(strings.NewReader(strings.Repeat("x", 1<<20+1)), "codex"); err == nil {
		t.Fatal("corpo acima de 1 MiB deveria ser recusado")
	}
	if got, err := readLatestBody(strings.NewReader("2.1.310\n"), "claude"); err != nil || string(got) != "2.1.310\n" {
		t.Fatalf("texto do Claude: %q, %v", got, err)
	}
	if _, err := readLatestBody(strings.NewReader(strings.Repeat("x", 5<<10)), "claude"); err == nil {
		t.Fatal("página de erro grande demais no endpoint do Claude deveria ser recusada")
	}
}
