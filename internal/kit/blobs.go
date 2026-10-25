package kit

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const blobDir = historyDir + "/blobs"
const blobKeepRecords = 5

var ErrPruned = errors.New("rollback alcança só as 5 gravações mais recentes; conteúdo anterior removido")

func blobValidHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

func blobPath(hash string) string {
	return blobDir + "/" + hash
}

func blobPut(root *os.Root, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	hash := digest(data)
	info, err := safePath(root, blobPath(hash))
	if err != nil {
		return err
	}
	if info != nil {
		return nil
	}
	sealed, err := protectHistory(data)
	if err != nil {
		return err
	}
	return replace(root, blobPath(hash), sealed, 0600)
}

func blobGet(root *os.Root, hash string) ([]byte, error) {
	if hash == digest(nil) {
		return []byte{}, nil
	}
	if !blobValidHash(hash) {
		return nil, fmt.Errorf("blob inválido: %q", hash)
	}
	data, _, exists, err := current(root, blobPath(hash))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPruned
	}
	plain, err := unprotectHistory(data)
	if err != nil {
		return nil, fmt.Errorf("blob inacessível: %w", err)
	}
	if digest(plain) != hash {
		return nil, fmt.Errorf("blob corrompido: %s", hash)
	}
	return plain, nil
}

func blobEntryBefore(root *os.Root, e entry) ([]byte, error) {
	if e.Before != nil {
		return e.Before, nil
	}
	return blobGet(root, e.BeforeHash)
}

func blobNewEntry(p string, before []byte, beforeMode fs.FileMode, existed bool, after []byte, afterMode fs.FileMode) entry {
	e := entry{Path: p, Existed: existed, AfterMode: afterMode, AfterHash: digest(after), AfterSize: len(after), After: after}
	if existed {
		if before == nil {
			before = []byte{}
		}
		e.Before = before
		e.BeforeMode = beforeMode
		e.BeforeHash = digest(before)
		e.BeforeSize = len(before)
	}
	return e
}

func blobRestoreSet(root *os.Root, r record) (map[string][]byte, error) {
	out := make(map[string][]byte)
	for _, e := range r.Entries {
		data, mode, exists, err := current(root, e.Path)
		if err != nil {
			return nil, err
		}
		if !e.Existed || matchesBefore(e, data, mode, exists) {
			continue
		}
		before, err := blobEntryBefore(root, e)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Path, err)
		}
		out[e.Path] = before
	}
	return out, nil
}

func blobMigrate(h *history) {
	for i := range h.Records {
		for j := range h.Records[i].Entries {
			e := &h.Records[i].Entries[j]
			if e.AfterHash == "" {
				e.AfterHash = digest(e.LegacyAfter)
				e.AfterSize = len(e.LegacyAfter)
			}
			if e.Existed && e.BeforeHash == "" {
				e.Before = e.LegacyBefore
				if e.Before == nil {
					e.Before = []byte{}
				}
				e.BeforeHash = digest(e.Before)
				e.BeforeSize = len(e.Before)
			}
			e.LegacyBefore, e.LegacyAfter = nil, nil
		}
	}
}

func blobPrune(h *history) bool {
	changed := false
	kept := h.Records[:0]
	for _, r := range h.Records {
		if r.RolledBack {
			changed = true
			continue
		}
		kept = append(kept, r)
	}
	h.Records = kept
	for i := range h.Records {
		r := &h.Records[i]
		if r.Pending || r.RollingBack || i >= len(h.Records)-blobKeepRecords {
			continue
		}
		if !r.Pruned {
			changed = true
		}
		r.Pruned = true
		for j := range r.Entries {
			r.Entries[j].Before, r.Entries[j].After = nil, nil
		}
	}
	return changed
}

func blobRemoveUnreferenced(root *os.Root, h history) error {
	keep := make(map[string]bool)
	for _, r := range h.Records {
		if r.Pruned {
			continue
		}
		for _, e := range r.Entries {
			if e.Existed {
				keep[e.BeforeHash] = true
			}
		}
	}
	info, err := safePath(root, blobDir)
	if err != nil || info == nil {
		return err
	}
	dir, err := root.Open(blobDir)
	if err != nil {
		return err
	}
	names, err := dir.Readdirnames(-1)
	dir.Close()
	if err != nil {
		return err
	}
	for _, name := range names {
		if !blobValidHash(name) || keep[name] {
			continue
		}
		if err := root.Remove(blobPath(name)); err != nil {
			return err
		}
	}
	return nil
}

func blobRetain(root *os.Root, h *history) error {
	changed := blobPrune(h)
	if changed {
		if err := saveHistory(root, *h); err != nil {
			return err
		}
	}
	return blobRemoveUnreferenced(root, *h)
}
