package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/Gex-Development/gex-dev-kit/internal/kit"
	udiff "github.com/aymanbagabas/go-udiff"
)

const (
	diffMasked        = "•••"
	diffCoreStart     = "<!-- início gex-dev-kit core -->\n"
	diffCoreEnd       = "<!-- fim gex-dev-kit core -->"
	diffHashLength    = 8
	diffStableSuffix  = "/.local/share/gex-dev-kit/bin/gex-dev-kit"
	diffStablePrefix  = ".local/share/gex-dev-kit/bin/"
	diffRemoveWarning = "os hooks passam a ignorar a verificação até nova instalação"
)

var (
	diffMaskLabel  = regexp.MustCompile(`(?i)(token|secret|password|passwd|api[_-]?key|authorization|bearer|webhook)([=:])\s*(?:(?:bearer|basic)\s+)?\S+`)
	diffMaskPrefix = regexp.MustCompile(`(ghp_|gho_|github_pat_|AKIA|ASIA|sk_live_|xox[baprs]-)[A-Za-z0-9_\-]*`)
	diffMaskURL    = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.\-]*://[^\s"'<>]*`)
	diffMaskLong   = regexp.MustCompile(`[A-Za-z0-9_+/=\-]{20,}`)
)

type diffEntry struct {
	Path      string
	Action    string
	Before    []byte
	After     []byte
	HasBefore bool
	HasAfter  bool
	Binary    bool
}

func maskLine(line string) (string, int) {
	count := 0
	line = diffMaskURL.ReplaceAllStringFunc(line, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || (u.Path == "" && u.RawQuery == "" && u.Fragment == "") {
			return raw
		}
		count++
		return u.Scheme + "://" + u.Host + "/" + diffMasked
	})
	line = diffMaskLabel.ReplaceAllStringFunc(line, func(raw string) string {
		count++
		parts := diffMaskLabel.FindStringSubmatch(raw)
		return parts[1] + parts[2] + diffMasked
	})
	line = diffMaskPrefix.ReplaceAllStringFunc(line, func(string) string {
		count++
		return diffMasked
	})
	line = diffMaskLong.ReplaceAllStringFunc(line, func(raw string) string {
		if !diffHasLetterAndDigit(raw) {
			return raw
		}
		count++
		return diffMasked
	})
	return line, count
}

func diffHasLetterAndDigit(s string) bool {
	letter, digit := false, false
	for _, r := range s {
		letter = letter || unicode.IsLetter(r)
		digit = digit || unicode.IsDigit(r)
	}
	return letter && digit
}

func diffShortHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:diffHashLength]
}

func diffCountLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	n := bytes.Count(data, []byte("\n"))
	if data[len(data)-1] != '\n' {
		n++
	}
	return n
}

func diffCoreBlock(data []byte) []byte {
	begin := bytes.Index(data, []byte(diffCoreStart))
	finish := bytes.Index(data, []byte(diffCoreEnd))
	if begin < 0 || finish < begin {
		return nil
	}
	return data[begin : finish+len(diffCoreEnd)]
}

func diffStableExe(p string) bool {
	p = strings.ReplaceAll(p, `\`, "/")
	return strings.HasSuffix(p, diffStableSuffix) || strings.HasSuffix(p, diffStableSuffix+".exe")
}

func diffKitHandler(h map[string]any) bool {
	cmd, _ := h["command"].(string)
	if i := strings.Index(cmd, " hook "); i >= 0 {
		return diffStableExe(strings.Trim(cmd[:i], `"`))
	}
	args, hasArgs := h["args"].([]any)
	if !hasArgs {
		return diffStableExe(cmd)
	}
	return len(args) > 0 && args[0] == "hook" && diffStableExe(cmd)
}

func diffHooksView(data []byte) ([]byte, int) {
	var doc map[string]any
	if len(bytes.TrimSpace(data)) == 0 || json.Unmarshal(data, &doc) != nil {
		return nil, 0
	}
	events, _ := doc["hooks"].(map[string]any)
	view := map[string]any{}
	third := 0
	for event, raw := range events {
		groups, _ := raw.([]any)
		var kept []any
		for _, g := range groups {
			group, _ := g.(map[string]any)
			handlers, _ := group["hooks"].([]any)
			var own []any
			for _, h := range handlers {
				handler, _ := h.(map[string]any)
				if diffKitHandler(handler) {
					own = append(own, handler)
				} else {
					third++
				}
			}
			if len(own) > 0 {
				kept = append(kept, map[string]any{"matcher": group["matcher"], "hooks": own})
			}
		}
		if len(kept) > 0 {
			view[event] = kept
		}
	}
	if len(view) == 0 {
		return nil, third
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(view); err != nil {
		return nil, third
	}
	return bytes.TrimRight(out.Bytes(), "\n"), third
}

func diffTextLines(before, after []byte, limit int) ([]DiffLine, int, int) {
	out := udiff.Unified("antes", "depois", string(before), string(after))
	if out == "" {
		return nil, 0, 0
	}
	rows := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(rows) >= 2 && strings.HasPrefix(rows[0], "--- ") && strings.HasPrefix(rows[1], "+++ ") {
		rows = rows[2:]
	}
	var lines []DiffLine
	for _, row := range rows {
		switch {
		case strings.HasPrefix(row, "@@"):
			lines = append(lines, DiffLine{Kind: '@', Text: row})
		case row != "" && strings.ContainsRune("+- ", rune(row[0])):
			lines = append(lines, DiffLine{Kind: row[0], Text: row[1:]})
		}
	}
	omitted := 0
	if limit < len(lines) {
		omitted = len(lines) - limit
		lines = lines[:limit]
	}
	masked := 0
	for i := range lines {
		if lines[i].Kind == '@' {
			continue
		}
		text, n := maskLine(lines[i].Text)
		lines[i].Text = text
		masked += n
	}
	return lines, omitted, masked
}

func diffBinarySummary(e diffEntry) string {
	switch {
	case e.HasBefore && e.HasAfter:
		return fmt.Sprintf("%d → %d bytes; sha256 %s → %s", len(e.Before), len(e.After), diffShortHash(e.Before), diffShortHash(e.After))
	case e.HasAfter:
		return fmt.Sprintf("novo, %d bytes; sha256 %s", len(e.After), diffShortHash(e.After))
	case e.HasBefore:
		return fmt.Sprintf("removido, %d bytes; sha256 %s", len(e.Before), diffShortHash(e.Before))
	}
	return ""
}

func diffOne(e diffEntry, limit int) FileDiff {
	d := FileDiff{Path: e.Path, Action: e.Action, Binary: e.Binary}
	if e.Binary {
		d.Summary = diffBinarySummary(e)
		return d
	}
	if !e.HasBefore && e.HasAfter {
		d.Summary = fmt.Sprintf("novo, %d linhas", diffCountLines(e.After))
		return d
	}
	if e.HasBefore && !e.HasAfter {
		d.Summary = fmt.Sprintf("removido, %d linhas", diffCountLines(e.Before))
	}
	before, after := e.Before, e.After
	switch e.Path {
	case ".claude/CLAUDE.md", ".codex/AGENTS.md":
		before, after = diffCoreBlock(before), diffCoreBlock(after)
	case ".claude/settings.json", ".codex/hooks.json":
		var beforeThird, afterThird int
		before, beforeThird = diffHooksView(before)
		after, afterThird = diffHooksView(after)
		third := afterThird
		if !e.HasAfter {
			third = beforeThird
		}
		if third > 0 {
			d.Summary = fmt.Sprintf("%d handlers de terceiros preservados, não exibidos", third)
		}
	}
	d.Lines, d.Omitted, d.Masked = diffTextLines(before, after, limit)
	return d
}

func buildDiffs(contents []kit.FileContent, limit int) []FileDiff {
	out := make([]FileDiff, 0, len(contents))
	for _, c := range contents {
		action := "criar"
		if c.Existed {
			action = "atualizar"
		}
		out = append(out, diffOne(diffEntry{Path: c.Path, Action: action, Before: c.Before, After: c.After, HasBefore: c.Existed, HasAfter: true, Binary: c.Binary}, limit))
	}
	return out
}

func buildRollbackDiffs(p kit.RollbackPlan, limit int) []FileDiff {
	out := make([]FileDiff, 0, len(p.Entries))
	for _, e := range p.Entries {
		if e.Action == "inalterado" {
			out = append(out, FileDiff{Path: e.Path, Action: e.Action, Binary: e.Binary})
			continue
		}
		d := diffOne(diffEntry{Path: e.Path, Action: e.Action, Before: e.Current, After: e.Restore, HasBefore: e.Current != nil, HasAfter: e.Action != "remover", Binary: e.Binary}, limit)
		if e.Action == "remover" && strings.HasPrefix(e.Path, diffStablePrefix) {
			if d.Summary == "" {
				d.Summary = diffRemoveWarning
			} else {
				d.Summary += "; " + diffRemoveWarning
			}
		}
		out = append(out, d)
	}
	return out
}
