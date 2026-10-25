package main

import (
	"bytes"
	"strings"
	"testing"
)

func invoke(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errors bytes.Buffer
	code := run(args, strings.NewReader(""), &out, &errors)
	return code, out.String(), errors.String()
}
