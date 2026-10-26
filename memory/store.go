package memory

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Record is a sourced fact that may be recalled within one project scope.
type Record struct {
	Version   int       `json:"version"`
	ID        string    `json:"id"`
	Scope     string    `json:"scope"`
	Kind      string    `json:"kind"`
	Content   string    `json:"content"`
	Sources   []string  `json:"sources"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct{ Dir string }

func (s Store) Remember(scope, kind, content string, sources []string) (Record, error) {
	scope, kind, content = strings.TrimSpace(scope), strings.TrimSpace(kind), strings.TrimSpace(content)
	if scope == "" || kind == "" || content == "" || len(sources) == 0 {
		return Record{}, errors.New("scope, kind, content, and source are required")
	}
	clean := make([]string, 0, len(sources))
	for _, source := range sources {
		source = strings.TrimSpace(source)
		parts := strings.SplitN(source, ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return Record{}, fmt.Errorf("source %q must have a type and reference, for example checkpoint:abc123", source)
		}
		clean = append(clean, source)
	}
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return Record{}, err
	}
	var idBytes [16]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return Record{}, err
	}
	r := Record{Version: 1, ID: hex.EncodeToString(idBytes[:]), Scope: scope, Kind: kind, Content: content, Sources: clean, CreatedAt: time.Now().UTC()}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return Record{}, err
	}
	name := filepath.Join(s.Dir, r.ID+".json")
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return Record{}, err
	}
	defer f.Close()
	if _, err = f.Write(append(data, '\n')); err != nil {
		os.Remove(name)
		return Record{}, err
	}
	if err = f.Sync(); err != nil {
		os.Remove(name)
		return Record{}, err
	}
	return r, nil
}

func (s Store) List(scope string) ([]Record, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Record
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !validID(id) {
			continue
		}
		r, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		if scope == "" || r.Scope == scope {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s Store) Search(scope, query string) ([]Record, error) {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if scope == "" || len(words) == 0 {
		return nil, errors.New("scope and query are required")
	}
	records, err := s.List(scope)
	if err != nil {
		return nil, err
	}
	var out []Record
	for _, r := range records {
		haystack := strings.ToLower(r.Content + " " + r.Kind + " " + strings.Join(r.Sources, " "))
		match := true
		for _, word := range words {
			if !strings.Contains(haystack, word) {
				match = false
				break
			}
		}
		if match {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s Store) Get(id string) (Record, error) {
	if !validID(id) {
		return Record{}, errors.New("invalid memory ID")
	}
	data, err := os.ReadFile(filepath.Join(s.Dir, id+".json"))
	if err != nil {
		return Record{}, err
	}
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return Record{}, err
	}
	if r.Version != 1 || r.ID != id {
		return Record{}, errors.New("unsupported or mismatched memory record")
	}
	return r, nil
}

func (s Store) Forget(id string) error {
	if !validID(id) {
		return errors.New("invalid memory ID")
	}
	return os.Remove(filepath.Join(s.Dir, id+".json"))
}

func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil && strings.ToLower(id) == id
}
