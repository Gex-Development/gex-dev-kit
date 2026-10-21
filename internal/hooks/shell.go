package hooks

import (
	"errors"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

var errShellSyntax = errors.New("comando com sintaxe não reconhecida")

const shellMaxDepth = 8

type Word struct {
	Value   string
	Dynamic bool
}

type Command struct {
	Words []Word
	Dir   string
}

type Script struct {
	Commands []Command
	Dynamic  bool
}

type shellHeredoc struct {
	delim  string
	quoted bool
	strip  bool
}

type shellParser struct {
	src      string
	pos      int
	depth    int
	script   Script
	words    []Word
	word     Word
	active   bool
	discard  bool
	cwd      string
	cwdDyn   bool
	heredocs []shellHeredoc
}

func ParseShell(command string) (Script, error) {
	return shellParse(command, 0)
}

func shellParse(src string, depth int) (Script, error) {
	if depth > shellMaxDepth {
		return Script{}, errShellSyntax
	}
	p := &shellParser{src: src, depth: depth}
	if err := p.run(); err != nil {
		return Script{}, err
	}
	return p.script, nil
}

func (p *shellParser) run() error {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		var err error
		switch {
		case c == '\\':
			p.backslash()
		case c == '\'':
			err = p.singleQuoted()
		case c == '"':
			err = p.doubleQuoted()
		case c == '$':
			err = p.expansion(false)
		case c == '`':
			err = p.backtick()
		case c == '#' && !p.active:
			p.skipComment()
		case c == ' ' || c == '\t' || c == '\r':
			p.flushWord()
			p.pos++
		case c == '\n':
			p.flushCommand()
			p.pos++
			err = p.readHeredocs()
		case c == '<' || c == '>' || c == '&' && p.peekAt(1) == '>':
			err = p.redirect()
		case strings.IndexByte(";|&()", c) >= 0:
			p.flushCommand()
			p.pos++
		default:
			p.addRune()
		}
		if err != nil {
			return err
		}
	}
	p.flushCommand()
	return nil
}

func (p *shellParser) peekAt(offset int) byte {
	if p.pos+offset < len(p.src) {
		return p.src[p.pos+offset]
	}
	return 0
}

func (p *shellParser) addRune() {
	_, size := utf8.DecodeRuneInString(p.src[p.pos:])
	p.word.Value += p.src[p.pos : p.pos+size]
	p.active = true
	p.pos += size
}

func (p *shellParser) backslash() {
	if p.pos+1 >= len(p.src) {
		p.word.Value += `\`
		p.active = true
		p.pos++
		return
	}
	if p.src[p.pos+1] == '\n' {
		p.pos += 2
		return
	}
	p.pos++
	p.addRune()
}

func (p *shellParser) skipComment() {
	for p.pos < len(p.src) && p.src[p.pos] != '\n' {
		p.pos++
	}
}

func (p *shellParser) skipBlanks() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *shellParser) flushWord() {
	if !p.active {
		return
	}
	if p.discard {
		p.discard = false
	} else {
		if p.word.Dynamic {
			p.script.Dynamic = true
		}
		p.words = append(p.words, p.word)
	}
	p.word = Word{}
	p.active = false
}

func (p *shellParser) flushCommand() {
	p.flushWord()
	p.discard = false
	if len(p.words) == 0 {
		return
	}
	cmd := Command{Words: p.words}
	p.words = nil
	p.trackDirectory(&cmd)
	p.script.Commands = append(p.script.Commands, cmd)
}

func (p *shellParser) trackDirectory(cmd *Command) {
	if !p.cwdDyn {
		cmd.Dir = p.cwd
	}
	first := cmd.Words[0]
	if first.Dynamic {
		return
	}
	switch first.Value {
	case "cd":
		if len(cmd.Words) == 2 {
			p.changeDirectory(cmd.Words[1])
		}
	case "git":
		if len(cmd.Words) >= 3 && cmd.Words[1].Value == "-C" {
			cmd.Dir = ""
			if !cmd.Words[2].Dynamic && !p.cwdDyn {
				cmd.Dir = shellJoinDir(p.cwd, cmd.Words[2].Value)
			}
		}
	}
}

func (p *shellParser) changeDirectory(w Word) {
	if w.Dynamic {
		p.cwdDyn = true
		p.cwd = ""
		return
	}
	if strings.HasPrefix(w.Value, "/") {
		p.cwdDyn = false
		p.cwd = w.Value
		return
	}
	if p.cwdDyn {
		return
	}
	p.cwd = shellJoinDir(p.cwd, w.Value)
}

func shellJoinDir(base, rel string) string {
	if base == "" || strings.HasPrefix(rel, "/") {
		return rel
	}
	return path.Join(base, rel)
}

func (p *shellParser) singleQuoted() error {
	end := strings.IndexByte(p.src[p.pos+1:], '\'')
	if end < 0 {
		return errShellSyntax
	}
	p.word.Value += p.src[p.pos+1 : p.pos+1+end]
	p.active = true
	p.pos += end + 2
	return nil
}

func (p *shellParser) doubleQuoted() error {
	p.active = true
	p.pos++
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == '"':
			p.pos++
			return nil
		case c == '\\' && p.pos+1 < len(p.src) && strings.IndexByte("$`\"\\\n", p.src[p.pos+1]) >= 0:
			if p.src[p.pos+1] != '\n' {
				p.word.Value += p.src[p.pos+1 : p.pos+2]
			}
			p.pos += 2
		case c == '$':
			if err := p.expansion(true); err != nil {
				return err
			}
		case c == '`':
			if err := p.backtick(); err != nil {
				return err
			}
		default:
			p.addRune()
		}
	}
	return errShellSyntax
}

func (p *shellParser) backtick() error {
	end, err := shellMatchBacktick(p.src, p.pos)
	if err != nil {
		return err
	}
	if err := p.appendSub(p.src[p.pos+1 : end]); err != nil {
		return err
	}
	p.dynamicRaw(p.src[p.pos : end+1])
	p.pos = end + 1
	return nil
}

func (p *shellParser) dynamicRaw(raw string) {
	p.word.Dynamic = true
	p.word.Value += raw
	p.active = true
}

func (p *shellParser) appendSub(inner string) error {
	sub, err := shellParse(inner, p.depth+1)
	if err != nil {
		return err
	}
	p.script.Commands = append(p.script.Commands, sub.Commands...)
	if sub.Dynamic {
		p.script.Dynamic = true
	}
	return nil
}

func (p *shellParser) expansion(inQuotes bool) error {
	next := p.peekAt(1)
	switch {
	case next == '\'' && !inQuotes:
		return p.ansiC()
	case next == '"' && !inQuotes:
		p.pos++
		return p.doubleQuoted()
	case next == '(':
		end, err := shellMatchParen(p.src, p.pos+1)
		if err != nil {
			return err
		}
		if err := p.appendSub(p.src[p.pos+2 : end]); err != nil {
			return err
		}
		p.dynamicRaw(p.src[p.pos : end+1])
		p.pos = end + 1
		return nil
	case next == '{':
		end := strings.IndexByte(p.src[p.pos:], '}')
		if end < 0 {
			return errShellSyntax
		}
		p.dynamicRaw(p.src[p.pos : p.pos+end+1])
		p.pos += end + 1
		return nil
	case next == '_' || shellIsLetter(next):
		end := p.pos + 1
		for end < len(p.src) && shellNameChar(p.src[end]) {
			end++
		}
		p.dynamicRaw(p.src[p.pos:end])
		p.pos = end
		return nil
	case shellIsDigit(next) || next != 0 && strings.IndexByte("?#$!*@-", next) >= 0:
		p.dynamicRaw(p.src[p.pos : p.pos+2])
		p.pos += 2
		return nil
	}
	p.word.Value += "$"
	p.active = true
	p.pos++
	return nil
}

func (p *shellParser) ansiC() error {
	var b strings.Builder
	i := p.pos + 2
	for i < len(p.src) {
		c := p.src[i]
		if c == '\'' {
			p.word.Value += b.String()
			p.active = true
			p.pos = i + 1
			return nil
		}
		if c == '\\' && i+1 < len(p.src) {
			decoded, size := shellEscape(p.src[i+1:])
			b.WriteString(decoded)
			i += 1 + size
			continue
		}
		b.WriteByte(c)
		i++
	}
	return errShellSyntax
}

func shellEscape(s string) (string, int) {
	switch s[0] {
	case 'n':
		return "\n", 1
	case 't':
		return "\t", 1
	case 'r':
		return "\r", 1
	case 'a':
		return "\a", 1
	case 'b':
		return "\b", 1
	case 'f':
		return "\f", 1
	case 'v':
		return "\v", 1
	case 'e', 'E':
		return string(rune(27)), 1
	case 'x':
		n := shellCountDigits(s[1:], 2, 16)
		if n == 0 {
			return "x", 1
		}
		v, _ := strconv.ParseUint(s[1:1+n], 16, 8)
		return string([]byte{byte(v)}), 1 + n
	}
	if s[0] >= '0' && s[0] <= '7' {
		n := shellCountDigits(s, 3, 8)
		v, _ := strconv.ParseUint(s[:n], 8, 16)
		return string([]byte{byte(v)}), n
	}
	return s[:1], 1
}

func shellCountDigits(s string, max, base int) int {
	n := 0
	for n < len(s) && n < max {
		c := s[n]
		ok := c >= '0' && c <= '7' || base == 16 && (c >= '8' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F')
		if !ok {
			break
		}
		n++
	}
	return n
}

func shellMatchParen(text string, open int) (int, error) {
	depth := 0
	for i := open; i < len(text); i++ {
		switch text[i] {
		case '\\':
			i++
		case '\'':
			end := strings.IndexByte(text[i+1:], '\'')
			if end < 0 {
				return 0, errShellSyntax
			}
			i += end + 1
		case '"':
			j := i + 1
			for j < len(text) && text[j] != '"' {
				if text[j] == '\\' {
					j++
				}
				j++
			}
			if j >= len(text) {
				return 0, errShellSyntax
			}
			i = j
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, errShellSyntax
}

func shellMatchBacktick(text string, open int) (int, error) {
	for i := open + 1; i < len(text); i++ {
		if text[i] == '\\' {
			i++
			continue
		}
		if text[i] == '`' {
			return i, nil
		}
	}
	return 0, errShellSyntax
}

func (p *shellParser) redirect() error {
	if p.active && !p.word.Dynamic && shellIsDigits(p.word.Value) {
		p.word = Word{}
		p.active = false
	} else {
		p.flushWord()
	}
	start := p.pos
	for p.pos < len(p.src) && strings.IndexByte("<>&|", p.src[p.pos]) >= 0 {
		p.pos++
	}
	if p.src[start:p.pos] == "<<" {
		strip := p.peekAt(0) == '-'
		if strip {
			p.pos++
		}
		h, err := p.heredocDelim(strip)
		if err != nil {
			return err
		}
		p.heredocs = append(p.heredocs, h)
		return nil
	}
	p.discard = true
	return nil
}

func (p *shellParser) heredocDelim(strip bool) (shellHeredoc, error) {
	p.skipBlanks()
	var b strings.Builder
	quoted := false
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if strings.IndexByte(" \t\r\n;|&<>()", c) >= 0 {
			break
		}
		switch c {
		case '\'', '"':
			quoted = true
			end := strings.IndexByte(p.src[p.pos+1:], c)
			if end < 0 {
				return shellHeredoc{}, errShellSyntax
			}
			b.WriteString(p.src[p.pos+1 : p.pos+1+end])
			p.pos += end + 2
		case '\\':
			quoted = true
			if p.pos+1 < len(p.src) {
				b.WriteByte(p.src[p.pos+1])
			}
			p.pos += 2
		default:
			b.WriteByte(c)
			p.pos++
		}
	}
	if p.pos > len(p.src) {
		p.pos = len(p.src)
	}
	return shellHeredoc{delim: b.String(), quoted: quoted, strip: strip}, nil
}

func (p *shellParser) readHeredocs() error {
	for _, h := range p.heredocs {
		bodyEnd, next := shellHeredocEnd(p.src, p.pos, h)
		if !h.quoted {
			if err := p.substitutionsIn(p.src[p.pos:bodyEnd]); err != nil {
				return err
			}
		}
		p.pos = next
	}
	p.heredocs = nil
	return nil
}

func shellHeredocEnd(src string, start int, h shellHeredoc) (int, int) {
	for lineStart := start; lineStart < len(src); {
		end, next := len(src), len(src)
		if i := strings.IndexByte(src[lineStart:], '\n'); i >= 0 {
			end = lineStart + i
			next = end + 1
		}
		line := strings.TrimSuffix(src[lineStart:end], "\r")
		if h.strip {
			line = strings.TrimLeft(line, "\t")
		}
		if line == h.delim {
			return lineStart, next
		}
		lineStart = next
	}
	return len(src), len(src)
}

func (p *shellParser) substitutionsIn(text string) error {
	for i := 0; i < len(text); i++ {
		switch {
		case text[i] == '$' && i+1 < len(text) && text[i+1] == '(':
			end, err := shellMatchParen(text, i+1)
			if err != nil {
				return err
			}
			if err := p.appendSub(text[i+2 : end]); err != nil {
				return err
			}
			i = end
		case text[i] == '`':
			end, err := shellMatchBacktick(text, i)
			if err != nil {
				return err
			}
			if err := p.appendSub(text[i+1 : end]); err != nil {
				return err
			}
			i = end
		}
	}
	return nil
}

func shellIsDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !shellIsDigit(s[i]) {
			return false
		}
	}
	return true
}

func shellIsLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func shellIsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func shellNameChar(c byte) bool {
	return c == '_' || shellIsLetter(c) || shellIsDigit(c)
}
