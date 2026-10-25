package kit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func blobTestRoot(t *testing.T) (string, *os.Root) {
	t.Helper()
	home := t.TempDir()
	root, err := os.OpenRoot(home)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return home, root
}

func TestBlobPutDeduplicatesAndStaysPrivate(t *testing.T) {
	home, root := blobTestRoot(t)
	data := []byte("conteudo anterior")
	for i := 0; i < 2; i++ {
		if err := blobPut(root, data); err != nil {
			t.Fatal(err)
		}
	}
	dir := filepath.Join(home, filepath.FromSlash(blobDir))
	names, err := os.ReadDir(dir)
	if err != nil || len(names) != 1 || names[0].Name() != digest(data) {
		t.Fatalf("blobs deduplicados: %v %v", names, err)
	}
	info, err := os.Stat(dir)
	if err != nil || !privateMode(info.Mode(), 0700) {
		t.Fatalf("pasta de blobs sem permissão privada: %v %v", info, err)
	}
	file, err := os.Stat(filepath.Join(dir, names[0].Name()))
	if err != nil || !privateMode(file.Mode(), 0600) {
		t.Fatalf("blob sem permissão privada: %v %v", file, err)
	}
}

func TestBlobGetRejectsInvalidAndPrunedHashes(t *testing.T) {
	_, root := blobTestRoot(t)
	if _, err := blobGet(root, "../../etc/passwd"); err == nil {
		t.Fatal("aceitou hash com caminho")
	}
	if _, err := blobGet(root, digest([]byte("removido"))); !errors.Is(err, ErrPruned) {
		t.Fatalf("blob ausente sem ErrPruned: %v", err)
	}
	if got, err := blobGet(root, digest(nil)); err != nil || len(got) != 0 {
		t.Fatalf("conteúdo vazio não lido: %q %v", got, err)
	}
}

func TestBlobPruneKeepsContentOnlyForFiveRecentRecords(t *testing.T) {
	var h history
	h.Records = append(h.Records, record{ID: "revertida", RolledBack: true})
	for i := 0; i < 6; i++ {
		content := []byte{byte(i)}
		h.Records = append(h.Records, record{ID: string(rune('a' + i)), Entries: []entry{{Path: "a.txt", Existed: true, BeforeHash: digest(content), Before: content}}})
	}
	if !blobPrune(&h) {
		t.Fatal("poda não reportou mudança")
	}
	if len(h.Records) != 6 {
		t.Fatalf("registro revertido não saiu do histórico: %d registros", len(h.Records))
	}
	for i, r := range h.Records {
		e := r.Entries[0]
		if i == 0 {
			if !r.Pruned || e.Before != nil || e.BeforeHash != digest([]byte{0}) {
				t.Fatalf("1ª gravação deveria ter o conteúdo podado e o hash mantido: %+v", r)
			}
			continue
		}
		if r.Pruned || e.Before == nil {
			t.Fatalf("gravação %d perdeu conteúdo dentro da janela: %+v", i, r)
		}
	}
}

func TestBlobMigrateConvertsLegacyInlineContent(t *testing.T) {
	h := history{Records: []record{{ID: "v02", Entries: []entry{{Path: "a.txt", Existed: true, LegacyBefore: []byte("old"), LegacyAfter: []byte("new")}}}}}
	blobMigrate(&h)
	e := h.Records[0].Entries[0]
	if e.BeforeHash != digest([]byte("old")) || e.AfterHash != digest([]byte("new")) || string(e.Before) != "old" || e.LegacyBefore != nil || e.LegacyAfter != nil {
		t.Fatalf("migração incorreta: %+v", e)
	}
}
