package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceScopedLifecycle(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	if _, err := s.Remember("repo", "decision", "Use device login", nil); err == nil {
		t.Fatal("accepted unsourced memory")
	}
	r, err := s.Remember("repo", "decision", "Use device login", []string{"checkpoint:abc123"})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(s.Dir, r.ID+".json"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("record permissions: %v, %v", info, err)
	}
	if got, err := s.Search("repo", "device login"); err != nil || len(got) != 1 || got[0].Sources[0] != "checkpoint:abc123" {
		t.Fatalf("source search: %#v, %v", got, err)
	}
	if got, err := s.Search("other", "device login"); err != nil || len(got) != 0 {
		t.Fatalf("scope leaked: %#v, %v", got, err)
	}
	if err := s.Forget(r.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(r.ID); !os.IsNotExist(err) {
		t.Fatalf("record still present: %v", err)
	}
}

func TestRejectTraversalAndCorruptRecord(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	if _, err := s.Get("../../secret"); err == nil {
		t.Fatal("accepted traversal ID")
	}
	id := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := os.WriteFile(filepath.Join(s.Dir, id+".json"), []byte(`{"version":1,"id":"different"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(id); err == nil {
		t.Fatal("accepted mismatched record ID")
	}
}
