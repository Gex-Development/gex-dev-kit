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
	"time"

	"github.com/Gex-Development/gex-dev-kit/internal/catalog"
)

var (
	ErrBusy       = errors.New("histórico ocupado")
	ErrChanged    = errors.New("arquivo alterado desde a prévia")
	ErrIncomplete = errors.New("operação anterior incompleta; execute rollback")
	ErrConflict   = errors.New("rollback abortado: arquivo alterado depois da instalação")
	ErrSymlink    = errors.New("link simbólico recusado")
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
	Path         string      `json:"path"`
	Existed      bool        `json:"existed"`
	BeforeMode   fs.FileMode `json:"before_mode,omitempty"`
	BeforeHash   string      `json:"before_hash,omitempty"`
	BeforeSize   int         `json:"before_size,omitempty"`
	AfterMode    fs.FileMode `json:"after_mode"`
	AfterHash    string      `json:"after_hash,omitempty"`
	AfterSize    int         `json:"after_size,omitempty"`
	LegacyBefore []byte      `json:"before,omitempty"`
	LegacyAfter  []byte      `json:"after,omitempty"`
	Before       []byte      `json:"-"`
	After        []byte      `json:"-"`
}

type record struct {
	ID          string  `json:"id"`
	Entries     []entry `json:"entries"`
	Pending     bool    `json:"pending,omitempty"`
	RollingBack bool    `json:"rolling_back,omitempty"`
	RolledBack  bool    `json:"rolled_back,omitempty"`
	Pruned      bool    `json:"pruned,omitempty"`
}

type history struct {
	Records []record `json:"records"`
}

const historyDir = ".local/share/gex-dev-kit/history"
const historyState = historyDir + "/state.json"
const lockName = historyDir + "/.lock"
const binDir = ".local/share/gex-dev-kit/bin/"
const binarySniff = 8192

type FileContent struct {
	Path    string
	Before  []byte
	After   []byte
	Existed bool
	Binary  bool
}

type RollbackEntry struct {
	Path    string
	Action  string
	Current []byte
	Restore []byte
	Binary  bool
}

type RollbackPlan struct {
	ID        string
	Entries   []RollbackEntry
	Conflicts []string
}

type LockInfo struct {
	PID     int
	Started time.Time
	Alive   bool
}

type lockFile struct {
	PID     int    `json:"pid"`
	Started string `json:"started"`
}

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
			return nil, fmt.Errorf("%w: %s", ErrSymlink, part)
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
		plan.entries = append(plan.entries, blobNewEntry(file.Path, saved, mode, existed, after, desiredMode))
		change := Change{Path: file.Path, Action: "criar", AfterHash: digest(after)}
		if existed {
			change.Action = "atualizar"
			change.BeforeHash = digest(saved)
		}
		plan.Changes = append(plan.Changes, change)
	}
	return plan, nil
}

func matchesBefore(e entry, data []byte, mode fs.FileMode, exists bool) bool {
	return exists == e.Existed && (!exists || digest(data) == e.BeforeHash && sameMode(mode, e.BeforeMode))
}

func matchesAfter(e entry, data []byte, mode fs.FileMode, exists bool) bool {
	return exists && digest(data) == e.AfterHash && sameMode(mode, e.AfterMode)
}

func rollbackMatches(e entry, data []byte, mode fs.FileMode, exists bool, interrupted bool) bool {
	return matchesAfter(e, data, mode, exists) || interrupted && matchesBefore(e, data, mode, exists)
}

func checkEntry(root *os.Root, e entry) error {
	data, mode, exists, err := current(root, e.Path)
	if err != nil {
		return err
	}
	if !matchesBefore(e, data, mode, exists) {
		return fmt.Errorf("%w: %s", ErrChanged, e.Path)
	}
	return nil
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
	blobMigrate(&h)
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
	for _, r := range h.Records {
		for _, e := range r.Entries {
			if e.Existed && e.Before != nil {
				if err := blobPut(root, e.Before); err != nil {
					return err
				}
			}
		}
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
	f, err := root.OpenFile(lockName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBusy, err)
	}
	data, err := json.Marshal(lockFile{PID: os.Getpid(), Started: time.Now().UTC().Format(time.RFC3339)})
	if err == nil {
		_, err = f.Write(data)
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		root.Remove(lockName)
		return nil, err
	}
	return func() { root.Remove(lockName) }, nil
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
			return "", ErrIncomplete
		}
	}
	for _, e := range plan.entries {
		if err := checkEntry(root, e); err != nil {
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
	if err := blobRetain(root, &h); err != nil {
		return "", fmt.Errorf("instalação concluída; limpeza do histórico falhou: %w", err)
	}
	return id, nil
}

func activeIndex(root *os.Root) (history, int, error) {
	info, err := safePath(root, historyState)
	if err != nil || info == nil {
		return history{}, -1, err
	}
	h, err := readHistory(root)
	if err != nil {
		return history{}, -1, err
	}
	for n := len(h.Records) - 1; n >= 0; n-- {
		if !h.Records[n].RolledBack {
			return h, n, nil
		}
	}
	return h, -1, nil
}

func rollbackConflicts(root *os.Root, r record) ([]string, error) {
	var conflicts []string
	for _, e := range r.Entries {
		data, mode, exists, err := current(root, e.Path)
		if err != nil {
			return nil, err
		}
		if !rollbackMatches(e, data, mode, exists, r.Pending || r.RollingBack) {
			conflicts = append(conflicts, e.Path)
		}
	}
	return conflicts, nil
}

func RollbackPreview(home string) (RollbackPlan, error) {
	root, err := openHome(home)
	if err != nil {
		return RollbackPlan{}, err
	}
	defer root.Close()
	h, index, err := activeIndex(root)
	if err != nil || index < 0 {
		return RollbackPlan{}, err
	}
	r := h.Records[index]
	conflicts, err := rollbackConflicts(root, r)
	if err != nil {
		return RollbackPlan{}, err
	}
	plan := RollbackPlan{ID: r.ID, Conflicts: conflicts}
	for _, e := range r.Entries {
		data, mode, exists, err := current(root, e.Path)
		if err != nil {
			return RollbackPlan{}, err
		}
		item := RollbackEntry{Path: e.Path, Action: "restaurar", Current: data, Binary: isStableBinary(e.Path) || binaryContent(data)}
		if matchesBefore(e, data, mode, exists) {
			item.Action = "inalterado"
		} else if !e.Existed {
			item.Action = "remover"
		} else {
			before, err := blobEntryBefore(root, e)
			if err != nil {
				return RollbackPlan{}, fmt.Errorf("%s: %w", e.Path, err)
			}
			item.Restore = before
			item.Binary = item.Binary || binaryContent(before)
		}
		plan.Entries = append(plan.Entries, item)
	}
	return plan, nil
}

func Rollback(home string) (string, error) {
	root, err := openHome(home)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if _, index, err := activeIndex(root); err != nil || index < 0 {
		return "", err
	}
	unlock, err := lock(root)
	if err != nil {
		return "", err
	}
	defer unlock()
	h, index, err := activeIndex(root)
	if err != nil || index < 0 {
		return "", err
	}
	r := &h.Records[index]
	conflicts, err := rollbackConflicts(root, *r)
	if err != nil {
		return "", err
	}
	if len(conflicts) > 0 {
		return "", fmt.Errorf("%w: %s", ErrConflict, strings.Join(conflicts, ", "))
	}
	restores, err := blobRestoreSet(root, *r)
	if err != nil {
		return "", err
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
		if matchesBefore(e, data, mode, exists) {
			continue
		}
		if e.Existed {
			if err := replace(root, e.Path, restores[e.Path], e.BeforeMode); err != nil {
				return "", err
			}
		} else if err := root.Remove(e.Path); err != nil {
			return "", err
		}
	}
	r.RolledBack = true
	r.RollingBack = false
	id := r.ID
	if err := saveHistory(root, h); err != nil {
		return "", err
	}
	if err := blobRetain(root, &h); err != nil {
		return "", fmt.Errorf("rollback concluído; limpeza do histórico falhou: %w", err)
	}
	return id, nil
}

func Doctor(home string) ([]Status, error) {
	root, err := openHome(home)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	status, err := lockStatus(root)
	if err != nil {
		return nil, err
	}
	info, err := safePath(root, historyState)
	if err != nil {
		return nil, err
	}
	if info == nil {
		if len(status) > 0 {
			return status, nil
		}
		return []Status{{Path: historyState, State: "sem-historico", Detail: "Nenhuma instalação registrada."}}, nil
	}
	h, err := readHistory(root)
	if err != nil {
		return nil, err
	}
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
		if errors.Is(err, ErrSymlink) {
			status = append(status, Status{Path: p, State: "bloqueado", Detail: "link simbólico; o kit não segue links."})
			continue
		}
		if err != nil {
			return nil, err
		}
		s := Status{Path: p, State: "ok", Detail: "Íntegro."}
		if !exists {
			s.State = "ausente"
			s.Detail = "Arquivo instalado ausente."
		} else if digest(data) != e.AfterHash || !sameMode(mode, e.AfterMode) {
			s.State = "alterado"
			s.Detail = "Arquivo foi alterado."
		}
		if p == ".codex/hooks.json" && s.State == "ok" && kitHookHandler(data) {
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

func lockStatus(root *os.Root) ([]Status, error) {
	info, present, err := readLock(root)
	if errors.Is(err, ErrSymlink) {
		return []Status{{Path: lockName, State: "bloqueado", Detail: "link simbólico; o kit não segue links."}}, nil
	}
	if err != nil || !present {
		return nil, err
	}
	if info.Alive {
		return []Status{{Path: lockName, State: "bloqueado", Detail: fmt.Sprintf("Outra execução do kit (PID %d) está ativa; aguarde.", info.PID)}}, nil
	}
	detail := fmt.Sprintf("O lock é de um processo que não existe mais (PID %d, desde %s). Remova %s e rode de novo.", info.PID, info.Started.Format(time.RFC3339), lockName)
	return []Status{{Path: lockName, State: "bloqueado", Detail: detail}}, nil
}

func readLock(root *os.Root) (LockInfo, bool, error) {
	info, err := safePath(root, lockName)
	if err != nil || info == nil {
		return LockInfo{}, false, err
	}
	data, _, exists, err := current(root, lockName)
	if err != nil || !exists {
		return LockInfo{}, false, err
	}
	var lf lockFile
	_ = json.Unmarshal(data, &lf)
	started, _ := time.Parse(time.RFC3339, lf.Started)
	return LockInfo{PID: lf.PID, Started: started, Alive: lockProcessAlive(lf.PID)}, true, nil
}

func Lock(home string) (LockInfo, bool, error) {
	root, err := openHome(home)
	if err != nil {
		return LockInfo{}, false, err
	}
	defer root.Close()
	return readLock(root)
}

func kitHookHandler(data []byte) bool {
	var doc any
	if json.Unmarshal(data, &doc) != nil {
		return false
	}
	return hasKitHookText(doc)
}

func hasKitHookText(v any) bool {
	switch x := v.(type) {
	case string:
		return strings.Contains(x, "gex-dev-kit") && strings.Contains(x, " hook")
	case []any:
		for _, item := range x {
			if hasKitHookText(item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range x {
			if hasKitHookText(item) {
				return true
			}
		}
	}
	return false
}

func (p Plan) Contents() []FileContent {
	out := make([]FileContent, 0, len(p.entries))
	for _, e := range p.entries {
		out = append(out, FileContent{Path: e.Path, Before: e.Before, After: e.After, Existed: e.Existed, Binary: isStableBinary(e.Path) || binaryContent(e.Before) || binaryContent(e.After)})
	}
	return out
}

func isStableBinary(p string) bool {
	return strings.HasPrefix(p, binDir)
}

func binaryContent(data []byte) bool {
	if len(data) > binarySniff {
		data = data[:binarySniff]
	}
	return bytes.IndexByte(data, 0) >= 0
}
