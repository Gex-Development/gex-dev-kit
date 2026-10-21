package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	ClaudeVersion = "2.1.306"
	CodexVersion  = "0.163.0"
)

const (
	installTimeout = 10 * time.Minute
	waitDelay      = 5 * time.Second
)

var probeTimeout = 5 * time.Second

var errCommandTimeout = errors.New("tempo esgotado")

var errOutputOpen = errors.New("processo filho manteve a saída aberta após o fim do comando")

type Change struct {
	Name    string
	Current string
	Target  string
	Command string
	Blocked bool
	Reason  string
	Manual  string
}

type record struct {
	Current  string `json:"current"`
	Previous string `json:"previous,omitempty"`
	Pending  bool   `json:"pending,omitempty"`
}
type history map[string]record

type product struct{ name, version, url string }

var products = []product{
	{"claude", ClaudeVersion, "https://claude.ai/install.sh"},
	{"codex", CodexVersion, "https://chatgpt.com/codex/install.sh"},
}
var exactVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func executable(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
func binPath(home, name string) string {
	if runtime.GOOS == "windows" && name == "codex" {
		return filepath.Join(home, "AppData", "Local", "Programs", "OpenAI", "Codex", "bin", executable(name))
	}
	return filepath.Join(home, ".local", "bin", executable(name))
}
func statePath(home string) string {
	return filepath.Join(home, ".local", "state", "gex-dev-kit", "agent-updates.json")
}
func validHome(home string) error {
	if home == "" || !filepath.IsAbs(home) || filepath.Clean(home) != home {
		return errors.New("HOME deve ser caminho absoluto normalizado")
	}
	info, err := os.Lstat(home)
	if err != nil {
		return fmt.Errorf("HOME: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("HOME não pode ser symlink")
	}
	if !info.IsDir() {
		return errors.New("HOME não é diretório")
	}
	return nil
}
func checkStateLinks(home string) error {
	for _, path := range []string{filepath.Join(home, ".local"), filepath.Join(home, ".local", "state"), filepath.Dir(statePath(home)), statePath(home)} {
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("histórico agentes contém symlink em %s", path)
		}
	}
	return nil
}

func load(home string) (history, error) {
	if err := checkStateLinks(home); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(statePath(home))
	if errors.Is(err, os.ErrNotExist) {
		return history{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("histórico agentes: %w", err)
	}
	var h history
	if err = json.Unmarshal(b, &h); err != nil || h == nil {
		return nil, errors.New("histórico agentes inválido; intervenção manual necessária")
	}
	for _, r := range h {
		if (r.Current != "" && !exactVersion.MatchString(r.Current)) || (r.Previous != "" && !exactVersion.MatchString(r.Previous)) || (r.Previous != "" && r.Current == "") {
			return nil, errors.New("histórico agentes inválido; intervenção manual necessária")
		}
	}
	return h, nil
}
func save(home string, h history) error {
	if err := checkStateLinks(home); err != nil {
		return err
	}
	path := statePath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".agents-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := checkStateLinks(home); err != nil {
		return err
	}
	return replaceFile(f.Name(), path)
}
func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
func pathEntries(name string) []string {
	var found []string
	seen := map[string]bool{}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, executable(name))
		info, err := os.Lstat(path)
		if err != nil || info.IsDir() {
			continue
		}
		key := path
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if !seen[key] {
			found = append(found, path)
			seen[key] = true
		}
	}
	return found
}
func manager(path string) string {
	lower := strings.ToLower(filepath.ToSlash(path))
	switch {
	case strings.Contains(lower, "/node_modules/"), strings.Contains(lower, "/npm/"):
		return "npm"
	case strings.Contains(lower, "/.bun/"):
		return "bun"
	case strings.Contains(lower, "/homebrew/"), strings.Contains(lower, "/cellar/"), strings.Contains(lower, "/caskroom/"):
		return "brew"
	case strings.Contains(lower, "/windowsapps/"), strings.Contains(lower, "/winget/"):
		return "winget"
	case strings.HasPrefix(lower, "/usr/bin/"):
		return "apt/dnf"
	default:
		return "não gerenciado pelo kit (brew/winget/apt/dnf/npm ou instalação própria)"
	}
}
func manual(name string) string {
	if name == "claude" {
		return "https://code.claude.com/docs/en/setup — verifique claude --version; npm install -g @anthropic-ai/claude-code@2.1.306; brew upgrade claude-code / winget upgrade Anthropic.ClaudeCode / sudo apt upgrade claude-code / sudo dnf upgrade claude-code NÃO garantem versão pinada"
	}
	return "https://github.com/openai/codex/releases/tag/rust-v0.163.0 — verifique codex --version; npm install -g @openai/codex@0.163.0; brew upgrade --cask codex NÃO garante versão pinada"
}
func installCommand(p product, version string) string {
	if runtime.GOOS == "windows" {
		if p.name == "claude" {
			return "PowerShell: & ([scriptblock]::Create((irm https://claude.ai/install.ps1))) " + version
		}
		return "PowerShell: $env:CODEX_NON_INTERACTIVE='1'; & ([scriptblock]::Create((irm https://chatgpt.com/codex/install.ps1))) -Release " + version
	}
	if p.name == "claude" {
		return "curl -fsSL https://claude.ai/install.sh | bash -s " + version
	}
	return "curl -fsSL https://chatgpt.com/codex/install.sh | CODEX_NON_INTERACTIVE=1 sh -s -- --release " + version
}
func probeReason(err error) string {
	if errors.Is(err, errCommandTimeout) {
		return "não verificado (tempo esgotado)"
	}
	return err.Error()
}
func observedVersion(home string, p product, path string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	output, err := runCommand(ctx, home, path, nil, "--version")
	if err != nil {
		return "", fmt.Errorf("%s --version: %w", p.name, err)
	}
	text := strings.ToLower(strings.TrimSpace(output))
	var pattern string
	if p.name == "claude" {
		pattern = `^(?:claude(?: code)?\s+)?(\d+\.\d+\.\d+)(?:\s+\(claude code\))?$`
	} else {
		pattern = `^codex(?:-cli)?\s+(\d+\.\d+\.\d+)$`
	}
	match := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if len(match) != 2 {
		return "", fmt.Errorf("%s --version fora do formato esperado", p.name)
	}
	return match[1], nil
}

var windowsUserPathReady = func(home, bin string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	output, err := runCommand(ctx, home, "powershell.exe", nil, "-NoProfile", "-NonInteractive", "-Command", "[Environment]::GetEnvironmentVariable('Path','User')")
	if err != nil {
		return false, err
	}
	for _, entry := range filepath.SplitList(strings.TrimSpace(output)) {
		if samePath(entry, bin) {
			return true, nil
		}
	}
	return false, nil
}

func Preview(home string) ([]Change, error) {
	if err := validHome(home); err != nil {
		return nil, err
	}
	h, err := load(home)
	if err != nil {
		return nil, err
	}
	changes := make([]Change, 0, len(products))
	for _, p := range products {
		r := h[p.name]
		c := Change{Name: p.name, Current: r.Current, Target: p.version, Command: installCommand(p, p.version), Manual: manual(p.name)}
		entries := pathEntries(p.name)
		target := binPath(home, p.name)
		info, statErr := os.Lstat(target)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return nil, fmt.Errorf("%s: %w", target, statErr)
		}
		exists := statErr == nil && !info.IsDir()
		switch {
		case r.Pending:
			c.Blocked = true
			c.Reason = "instalação anterior interrompida; verificar binário e histórico manualmente"
		case len(entries) > 1:
			c.Blocked = true
			c.Reason = "múltiplos executáveis no PATH; resolver colisão antes da instalação"
		case len(entries) == 1 && !samePath(entries[0], target):
			resolved, err := filepath.EvalSymlinks(entries[0])
			if err != nil {
				return nil, fmt.Errorf("%s: %w", entries[0], err)
			}
			c.Blocked = true
			c.Reason = manager(resolved) + " em " + entries[0] + "; não alterar instalação alheia"
		case exists && r.Current == "":
			c.Blocked = true
			c.Reason = "binário não gerenciado pelo kit em " + target
		case r.Current != "" && !exists:
			c.Blocked = true
			c.Reason = "binário gerenciado ausente; histórico divergente"
		case len(entries) == 0:
			c.Blocked = true
			c.Reason = filepath.Dir(target) + " não está no PATH; não instalar binário invisível"
		case !samePath(entries[0], target):
			c.Blocked = true
			c.Reason = "PATH não seleciona binário do kit"
		default:
			if exists {
				version, e := observedVersion(home, p, target)
				if e != nil {
					c.Blocked = true
					c.Reason = probeReason(e)
				} else if version != r.Current {
					c.Blocked = true
					c.Reason = "versão instalada " + version + " difere do histórico " + r.Current + " (possível atualização automática)"
				}
			}
		}
		if !exists && r.Current == "" && len(entries) == 0 && !r.Pending {
			onPath := false
			for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
				if samePath(dir, filepath.Dir(target)) {
					onPath = true
				}
			}
			if onPath {
				c.Blocked = false
				c.Reason = ""
			}
		}
		if runtime.GOOS == "windows" && p.name == "codex" && !c.Blocked {
			ready, probeErr := windowsUserPathReady(home, filepath.Dir(target))
			switch {
			case errors.Is(probeErr, errCommandTimeout):
				c.Blocked = true
				c.Reason = probeReason(probeErr)
			case probeErr != nil || !ready:
				c.Blocked = true
				c.Reason = "Codex alteraria PATH do usuário; inclua ~/.local/bin no PATH de usuário antes de instalar"
			}
		}
		changes = append(changes, c)
	}
	return changes, nil
}

func requireLock(home string) (func(), error) {
	if err := checkStateLinks(home); err != nil {
		return nil, err
	}
	dir := filepath.Dir(statePath(home))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	lock := filepath.Join(dir, "agent-updates.lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		return nil, fmt.Errorf("outra atualização ou lock pendente: %w", err)
	}
	unlock := func() {
		_ = os.Remove(filepath.Join(lock, "owner.json"))
		_ = os.Remove(lock)
	}
	if err := writeLockOwner(lock); err != nil {
		unlock()
		return nil, err
	}
	return unlock, nil
}

type lockOwner struct {
	PID     int    `json:"pid"`
	Started string `json:"started"`
}

func writeLockOwner(lock string) error {
	b, err := json.Marshal(lockOwner{PID: os.Getpid(), Started: time.Now().Format(time.RFC3339)})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(lock, "owner.json"), b, 0600)
}
func Apply(home string, changes []Change) error {
	if err := validHome(home); err != nil {
		return err
	}
	if runtime.GOOS != "windows" && os.Geteuid() == 0 {
		return errors.New("instalador standalone não pode rodar como root")
	}
	unlock, err := requireLock(home)
	if err != nil {
		return err
	}
	defer unlock()
	fresh, err := Preview(home)
	if err != nil {
		return err
	}
	h, err := load(home)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, c := range changes {
		if seen[c.Name] {
			return fmt.Errorf("mudança duplicada: %s", c.Name)
		}
		seen[c.Name] = true
		index := -1
		for i, p := range products {
			if p.name == c.Name {
				index = i
				break
			}
		}
		if index < 0 || c != fresh[index] {
			return fmt.Errorf("prévia desatualizada ou inválida: %s", c.Name)
		}
		if c.Blocked {
			return fmt.Errorf("%s bloqueado: %s; %s", c.Name, c.Reason, c.Manual)
		}
	}
	for _, c := range changes {
		if c.Current == c.Target {
			continue
		}
		p := products[0]
		if c.Name == "codex" {
			p = products[1]
		}
		r := h[p.name]
		r.Pending = true
		h[p.name] = r
		if err := save(home, h); err != nil {
			return err
		}
		if err := install(home, p, p.version); err != nil {
			return fmt.Errorf("%s instalação incompleta, verificar manualmente: %w", p.name, err)
		}
		r.Pending = false
		r.Previous = r.Current
		r.Current = p.version
		h[p.name] = r
		if err := save(home, h); err != nil {
			return fmt.Errorf("%s instalado mas histórico não persistiu: %w", p.name, err)
		}
	}
	return nil
}

var fetchScript = func(url string) ([]byte, error) {
	response, err := (&http.Client{Timeout: 30 * time.Second}).Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	allowed := map[string]bool{"downloads.claude.ai": true, "releases.openai.com": true}
	if response.Request.URL.Scheme != "https" || !allowed[response.Request.URL.Hostname()] {
		return nil, errors.New("redirecionamento do instalador fora do domínio oficial")
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, (2<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 2<<20 {
		return nil, errors.New("instalador excede 2 MiB")
	}
	return b, nil
}

func safeEnvironment(home string) []string {
	allow := map[string]bool{"PATH": true, "SYSTEMROOT": true, "WINDIR": true, "COMSPEC": true, "PATHEXT": true, "PROCESSOR_ARCHITECTURE": true, "PROCESSOR_ARCHITEW6432": true, "LANG": true, "LC_ALL": true, "SSL_CERT_FILE": true, "SSL_CERT_DIR": true}
	env := make([]string, 0, 20)
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if allow[strings.ToUpper(key)] {
			env = append(env, v)
		}
	}
	env = append(env, "HOME="+home, "USERPROFILE="+home, "LOCALAPPDATA="+filepath.Join(home, "AppData", "Local"), "APPDATA="+filepath.Join(home, "AppData", "Roaming"), "TMPDIR="+home, "TMP="+home, "TEMP="+home, "CODEX_HOME="+filepath.Join(home, ".codex"), "CODEX_NON_INTERACTIVE=1")
	if runtime.GOOS != "windows" {
		env = append(env, "CODEX_INSTALL_DIR="+filepath.Join(home, ".local", "bin"))
	}
	return env
}

var runCommand = func(ctx context.Context, home, path string, input []byte, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = safeEnvironment(home)
	cmd.Stdin = bytes.NewReader(input)
	cmd.WaitDelay = waitDelay
	var out bytes.Buffer
	cmd.Stdout = &out
	group, err := procNewGroup(cmd)
	if err != nil {
		return "", err
	}
	defer group.close()
	cmd.Cancel = func() error { return group.kill(cmd) }
	if err = runGroup(cmd, group); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", errCommandTimeout
		}
		if errors.Is(err, exec.ErrWaitDelay) {
			return "", errOutputOpen
		}
		return "", err
	}
	return out.String(), nil
}

func runGroup(cmd *exec.Cmd, group *procGroup) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := group.adopt(cmd); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	err := cmd.Wait()
	_ = group.kill(cmd)
	return err
}

func install(home string, p product, version string) error {
	if runtime.GOOS != "windows" && os.Geteuid() == 0 {
		return errors.New("instalador standalone não pode rodar como root")
	}
	if !exactVersion.MatchString(version) {
		return errors.New("versão inválida")
	}
	url := p.url
	if runtime.GOOS == "windows" {
		url = strings.TrimSuffix(url, ".sh") + ".ps1"
	}
	script, err := fetchScript(url)
	if err != nil {
		return fmt.Errorf("baixar instalador oficial %s: %w", url, err)
	}
	var path string
	var args []string
	var input []byte
	if runtime.GOOS == "windows" {
		f, e := os.CreateTemp(home, "gex-agent-*.ps1")
		if e != nil {
			return e
		}
		path = f.Name()
		defer os.Remove(path)
		if _, e = f.Write(script); e != nil {
			f.Close()
			return e
		}
		if e = f.Close(); e != nil {
			return e
		}
		input = nil
		args = []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path}
		if p.name == "claude" {
			args = append(args, version)
		} else {
			args = append(args, "-Release", version)
		}
		path = "powershell.exe"
	} else {
		input = script
		if p.name == "claude" {
			path = "bash"
			args = []string{"-s", "--", version}
		} else {
			path = "sh"
			args = []string{"-s", "--", "--release", version}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()
	if _, err = runCommand(ctx, home, path, input, args...); err != nil {
		return fmt.Errorf("instalador oficial: %w", err)
	}
	actual, err := observedVersion(home, p, binPath(home, p.name))
	if err != nil {
		return err
	}
	if actual != version {
		return fmt.Errorf("%s --version = %s, esperado %s", p.name, actual, version)
	}
	return nil
}
func Rollback(home string) error {
	if err := validHome(home); err != nil {
		return err
	}
	if runtime.GOOS != "windows" && os.Geteuid() == 0 {
		return errors.New("instalador standalone não pode rodar como root")
	}
	unlock, err := requireLock(home)
	if err != nil {
		return err
	}
	defer unlock()
	changes, err := Preview(home)
	if err != nil {
		return err
	}
	h, err := load(home)
	if err != nil {
		return err
	}
	for i, p := range products {
		r := h[p.name]
		if r.Current != "" && r.Previous == "" {
			return fmt.Errorf("rollback %s indisponível: instalação inicial sem versão anterior; desinstalação automática insegura", p.name)
		}
		if r.Previous != "" && changes[i].Blocked {
			return fmt.Errorf("%s bloqueado: %s", p.name, changes[i].Reason)
		}
	}
	restored := false
	for _, p := range products {
		r := h[p.name]
		if r.Previous == "" {
			continue
		}
		restored = true
		prior := r.Previous
		r.Pending = true
		h[p.name] = r
		if err := save(home, h); err != nil {
			return err
		}
		if err := install(home, p, prior); err != nil {
			return fmt.Errorf("rollback %s incompleto: %w", p.name, err)
		}
		r.Pending = false
		r.Current = prior
		r.Previous = ""
		h[p.name] = r
		if err := save(home, h); err != nil {
			return err
		}
	}
	if !restored {
		return errors.New("nenhuma instalação anterior gerenciada pelo kit para restaurar")
	}
	return nil
}
