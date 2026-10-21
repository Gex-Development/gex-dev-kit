package kit

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

type Plan struct {
	Changes []Change
	home    string
	source  string
	entries []entry
}

type Change struct {
	Path       string
	Action     string
	BeforeHash string
	AfterHash  string
}

type Status struct {
	Path   string
	State  string
	Detail string
}

type entry struct {
	Path       string      `json:"path"`
	Before     []byte      `json:"before,omitempty"`
	BeforeMode fs.FileMode `json:"before_mode,omitempty"`
	Existed    bool        `json:"existed"`
	After      []byte      `json:"after"`
	AfterMode  fs.FileMode `json:"after_mode"`
}

type record struct {
	ID          string  `json:"id"`
	Entries     []entry `json:"entries"`
	Pending     bool    `json:"pending,omitempty"`
	RollingBack bool    `json:"rolling_back,omitempty"`
	RolledBack  bool    `json:"rolled_back,omitempty"`
}

type history struct {
	Records []record `json:"records"`
}

const historyDir = ".local/share/gex-dev-kit/history"
const historyState = historyDir + "/state.json"

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func canonicalHome(home string) (string, error) {
	if home == "" || !filepath.IsAbs(home) || filepath.Clean(home) != home {
		return "", errors.New("HOME deve ser absoluto e normalizado")
	}
	resolved, err := filepath.EvalSymlinks(home)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("HOME não é diretório")
	}
	return resolved, nil
}

func validPath(p string) error {
	if p == "" || path.IsAbs(p) || filepath.IsAbs(p) || filepath.VolumeName(p) != "" || strings.Contains(p, "\\") || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") || p == "." {
		return fmt.Errorf("caminho inválido: %q", p)
	}
	return nil
}

func safePath(root *os.Root, p string) (fs.FileInfo, error) {
	if err := validPath(p); err != nil {
		return nil, err
	}
	parts := strings.Split(p, "/")
	for n := range parts {
		part := strings.Join(parts[:n+1], "/")
		info, err := root.Lstat(part)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("link simbólico recusado: %s", part)
		}
		if n < len(parts)-1 && !info.IsDir() {
			return nil, fmt.Errorf("diretório inválido: %s", part)
		}
		if n == len(parts)-1 {
			return info, nil
		}
	}
	return nil, nil
}

func openHome(home string) (*os.Root, error) {
	resolved, err := canonicalHome(home)
	if err != nil {
		return nil, err
	}
	return os.OpenRoot(resolved)
}

func current(root *os.Root, p string) ([]byte, fs.FileMode, bool, error) {
	info, err := safePath(root, p)
	if err != nil || info == nil {
		return nil, 0, false, err
	}
	if !info.Mode().IsRegular() {
		return nil, 0, false, fmt.Errorf("arquivo não regular: %s", p)
	}
	data, err := root.ReadFile(p)
	return data, info.Mode(), true, err
}

func Preview(home string, files []catalog.File) (Plan, error) {
	resolved, err := canonicalHome(home)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{home: resolved, source: home}
	root, err := os.OpenRoot(resolved)
	if err != nil {
		return Plan{}, err
	}
	defer root.Close()
	seen := make(map[string]bool, len(files))
	for _, file := range files {
		if err := validPath(file.Path); err != nil {
			return Plan{}, err
		}
		if file.Path == historyDir || strings.HasPrefix(file.Path, historyDir+"/") {
			return Plan{}, errors.New("arquivo reservado ao histórico")
		}
		if seen[file.Path] {
			return Plan{}, fmt.Errorf("caminho duplicado: %s", file.Path)
		}
		seen[file.Path] = true
		if !file.Mode.IsRegular() {
			return Plan{}, fmt.Errorf("modo inválido: %s", file.Path)
		}
		before, mode, existed, err := current(root, file.Path)
		if err != nil {
			return Plan{}, err
		}
		after := append([]byte(nil), file.Data...)
		desiredMode := file.Mode.Perm() | file.Mode&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky)
		if file.Path == ".claude/settings.json" || file.Path == ".codex/hooks.json" {
			desiredMode &= 0600
		}
		if existed && bytes.Equal(before, after) && sameMode(mode, desiredMode) {
			continue
		}
		saved := before
		plan.entries = append(plan.entries, entry{Path: file.Path, Before: saved, BeforeMode: mode, Existed: existed, After: after, AfterMode: desiredMode})
		change := Change{Path: file.Path, Action: "criar", AfterHash: digest(after)}
		if existed {
			change.Action = "atualizar"
			change.BeforeHash = digest(saved)
		}
		plan.Changes = append(plan.Changes, change)
	}
	return plan, nil
}

func checkEntry(root *os.Root, e entry, after bool, allowBefore bool) error {
	data, mode, exists, err := current(root, e.Path)
	if err != nil {
		return err
	}
	if after && exists && digest(data) == digest(e.After) && sameMode(mode, e.AfterMode) {
		return nil
	}
	if allowBefore && exists == e.Existed && (!exists || digest(data) == digest(e.Before) && sameMode(mode, e.BeforeMode)) {
		return nil
	}
	if !after && exists == e.Existed && (!exists || digest(data) == digest(e.Before) && sameMode(mode, e.BeforeMode)) {
		return nil
	}
	return fmt.Errorf("arquivo alterado desde a prévia: %s", e.Path)
}

func readHistory(root *os.Root) (history, error) {
	data, mode, exists, err := current(root, historyState)
	if err != nil || !exists {
		return history{}, err
	}
	if !privateMode(mode, 0600) {
		return history{}, errors.New("histórico sem permissão privada")
	}
	data, err = unprotectHistory(data)
	if err != nil {
		return history{}, fmt.Errorf("histórico inacessível: %w", err)
	}
	var h history
	if err := json.Unmarshal(data, &h); err != nil {
		return history{}, err
	}
	return h, nil
}

func randomID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func ensureParents(root *os.Root, p string) error {
	dir := path.Dir(p)
	if dir == "." {
		return nil
	}
	parts := strings.Split(dir, "/")
	for n := range parts {
		part := strings.Join(parts[:n+1], "/")
		info, err := safePath(root, part)
		if err != nil {
			return err
		}
		if info == nil {
			if err := root.Mkdir(part, 0700); err != nil {
				return err
			}
		} else if !info.IsDir() {
			return fmt.Errorf("diretório inválido: %s", part)
		}
	}
	return nil
}

func replace(root *os.Root, p string, data []byte, mode fs.FileMode) error {
	if _, err := safePath(root, p); err != nil {
		return err
	}
	if err := ensureParents(root, p); err != nil {
		return err
	}
	id, err := randomID()
	if err != nil {
		return err
	}
	temp := path.Join(path.Dir(p), ".gex-dev-kit-"+id)
	f, err := root.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(temp)
	if _, err = f.Write(data); err == nil {
		err = f.Chmod(mode)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return root.Rename(temp, p)
}

func saveHistory(root *os.Root, h history) error {
	if err := ensureParents(root, historyState); err != nil {
		return err
	}
	info, err := safePath(root, historyDir)
	if err != nil {
		return err
	}
	if info == nil || !privateMode(info.Mode(), 0700) {
		return errors.New("histórico sem permissão privada")
	}
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	data, err = protectHistory(data)
	if err != nil {
		return err
	}
	return replace(root, historyState, data, 0600)
}

func lock(root *os.Root) (func(), error) {
	if err := ensureParents(root, historyState); err != nil {
		return nil, err
	}
	info, err := safePath(root, historyDir)
	if err != nil {
		return nil, err
	}
	if !privateMode(info.Mode(), 0700) {
		return nil, errors.New("histórico sem permissão privada")
	}
	f, err := root.OpenFile(historyDir+"/.lock", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("histórico ocupado: %w", err)
	}
	if err := f.Close(); err != nil {
		root.Remove(historyDir + "/.lock")
		return nil, err
	}
	return func() { root.Remove(historyDir + "/.lock") }, nil
}

func Apply(plan Plan) (string, error) {
	if plan.home == "" {
		return "", errors.New("prévia obrigatória")
	}
	resolved, err := canonicalHome(plan.source)
	if err != nil {
		return "", err
	}
	if resolved != plan.home {
		return "", errors.New("HOME mudou desde a prévia")
	}
	root, err := os.OpenRoot(plan.home)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if len(plan.entries) == 0 {
		return "", nil
	}
	unlock, err := lock(root)
	if err != nil {
		return "", err
	}
	defer unlock()
	h, err := readHistory(root)
	if err != nil {
		return "", err
	}
	for _, r := range h.Records {
		if (r.Pending || r.RollingBack) && !r.RolledBack {
			return "", errors.New("operação anterior incompleta; execute rollback")
		}
	}
	for _, e := range plan.entries {
		if err := checkEntry(root, e, false, false); err != nil {
			return "", err
		}
	}
	id, err := randomID()
	if err != nil {
		return "", err
	}
	h.Records = append(h.Records, record{ID: id, Entries: plan.entries, Pending: true})
	if err := saveHistory(root, h); err != nil {
		return "", err
	}
	for _, e := range plan.entries {
		if err := replace(root, e.Path, e.After, e.AfterMode); err != nil {
			return "", fmt.Errorf("instalação incompleta; execute rollback: %w", err)
		}
	}
	h.Records[len(h.Records)-1].Pending = false
	if err := saveHistory(root, h); err != nil {
		return "", fmt.Errorf("instalação concluída, histórico pendente; execute rollback: %w", err)
	}
	return id, nil
}

func Rollback(home string) (string, error) {
	root, err := openHome(home)
	if err != nil {
		return "", err
	}
	defer root.Close()
	info, err := safePath(root, historyState)
	if err != nil {
		return "", err
	}
	if info == nil {
		return "", nil
	}
	unlock, err := lock(root)
	if err != nil {
		return "", err
	}
	defer unlock()
	h, err := readHistory(root)
	if err != nil {
		return "", err
	}
	index := -1
	for n := len(h.Records) - 1; n >= 0; n-- {
		if !h.Records[n].RolledBack {
			index = n
			break
		}
	}
	if index < 0 {
		return "", nil
	}
	r := &h.Records[index]
	for _, e := range r.Entries {
		if err := checkEntry(root, e, true, r.Pending || r.RollingBack); err != nil {
			return "", err
		}
	}
	r.RollingBack = true
	if err := saveHistory(root, h); err != nil {
		return "", err
	}
	for _, e := range r.Entries {
		data, mode, exists, err := current(root, e.Path)
		if err != nil {
			return "", err
		}
		if exists == e.Existed && (!exists || digest(data) == digest(e.Before) && sameMode(mode, e.BeforeMode)) {
			continue
		}
		if e.Existed {
			if err := replace(root, e.Path, e.Before, e.BeforeMode); err != nil {
				return "", err
			}
		} else if err := root.Remove(e.Path); err != nil {
			return "", err
		}
	}
	r.RolledBack = true
	r.RollingBack = false
	if err := saveHistory(root, h); err != nil {
		return "", err
	}
	return r.ID, nil
}

func Doctor(home string) ([]Status, error) {
	root, err := openHome(home)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := safePath(root, historyState)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return []Status{{Path: historyState, State: "sem-historico", Detail: "Nenhuma instalação registrada."}}, nil
	}
	h, err := readHistory(root)
	if err != nil {
		return nil, err
	}
	status := []Status{}
	latest := make(map[string]entry)
	for _, r := range h.Records {
		if r.RolledBack {
			continue
		}
		for _, e := range r.Entries {
			latest[e.Path] = e
		}
		if r.Pending || r.RollingBack {
			status = append(status, Status{Path: historyState, State: "incompleto", Detail: "Operação interrompida; execute rollback."})
		}
	}
	for p, e := range latest {
		data, mode, exists, err := current(root, p)
		if err != nil {
			return nil, err
		}
		s := Status{Path: p, State: "ok", Detail: "Íntegro."}
		if !exists {
			s.State = "ausente"
			s.Detail = "Arquivo instalado ausente."
		} else if digest(data) != digest(e.After) || !sameMode(mode, e.AfterMode) {
			s.State = "alterado"
			s.Detail = "Arquivo foi alterado."
		}
		if p == ".codex/hooks.json" && s.State == "ok" && strings.Contains(string(e.After), "hook commit-msg") && strings.Contains(string(e.After), "gex-dev-kit") {
			s.State = "pendente"
			s.Detail = "Confirme confiança do hook no Codex com /hooks; estado de confiança não verificável aqui."
		}
		status = append(status, s)
	}
	sort.Slice(status, func(i, j int) bool { return status[i].Path < status[j].Path })
	if len(status) == 0 {
		status = append(status, Status{Path: historyState, State: "sem-historico", Detail: "Nenhuma instalação ativa."})
	}
	return status, nil
}
