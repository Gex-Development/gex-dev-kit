package app

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type agentLatestTransport struct{}

func (agentLatestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := "2.1.306"
	if req.URL.Hostname() == "releases.openai.com" {
		body = `{"tag_name":"rust-v0.163.0"}`
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Request: req, Header: make(http.Header)}, nil
}

func TestMain(m *testing.M) {
	http.DefaultTransport = agentLatestTransport{}
	os.Exit(m.Run())
}
