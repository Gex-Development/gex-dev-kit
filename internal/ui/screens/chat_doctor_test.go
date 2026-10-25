package screens

import (
	"strings"
	"testing"

	"github.com/Gex-Development/gex-dev-kit/internal/app"
)

func TestChatDoctorReportsAttentionAndExitCode(t *testing.T) {
	m := chatTestModel(false)
	m.deps.Doctor = func() (app.Report, error) {
		return app.Report{Root: "/tmp/repo", Items: []app.Status{{Path: "changed.txt", State: "alterado"}}}, nil
	}
	m = m.flowDoctor()
	m = chatSetupQuestion(t, m, "O que eu faço?")
	if !strings.Contains(chatTestTranscript(m), "Encontrei 1 item(ns) com atenção") || m.exitCode != 1 {
		t.Fatalf("doctor precisa relatar atenção e código 1: %d %s", m.exitCode, chatTestTranscript(m))
	}
}
