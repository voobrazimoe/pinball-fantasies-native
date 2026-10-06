package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"pinballfantasies/internal/tablelogic"
)

type Score struct {
	Digits tablelogic.Decimal
	Name   [3]byte
}
type Scores [4]Score

func Defaults(table int) Scores {
	values := [4]uint64{50_000_000, 25_000_000, 10_000_000, 5_000_000}
	names := [4]string{"TSP", "ICE", "ANY", "J L"}
	if table != 1 {
		values = [4]uint64{100_000_000, 50_000_000, 25_000_000, 10_000_000}
		switch table {
		case 2:
			names = [4]string{"TSP", "J L", "ICE", "ANY"}
		case 3:
			names = [4]string{"TSP", "ANY", "J L", "ICE"}
		case 4:
			names = [4]string{"TSP", "ICE", "ANY", "J L"}
		}
	}
	var s Scores
	for i := range s {
		s[i].Digits = tablelogic.Number(values[i])
		copy(s[i].Name[:], names[i])
	}
	return s
}

// Rank is the original most-significant-first strict comparison; ties stay below
// an equal entry and may still beat a lower entry.
func (s Scores) Rank(d tablelogic.Decimal) int {
	for i, e := range s {
		if d.Uint64() > e.Digits.Uint64() {
			return i
		}
	}
	return -1
}
func (s *Scores) Insert(rank int, d tablelogic.Decimal, name [3]byte) {
	copy(s[rank+1:], s[rank:3])
	s[rank] = Score{d, name}
}
func (s Scores) MarshalBinary() []byte {
	b := make([]byte, 64)
	for i, e := range s {
		copy(b[i*16:], e.Digits[:])
		copy(b[i*16+12:], e.Name[:])
	}
	return b
}
func DecodeScores(b []byte) (s Scores, err error) {
	if len(b) != 64 {
		return s, fmt.Errorf("high scores must contain four 16-byte records")
	}
	for i := range s {
		for j, v := range b[i*16 : i*16+12] {
			if v > 9 {
				return s, fmt.Errorf("invalid decimal digit")
			}
			s[i].Digits[j] = v
		}
		copy(s[i].Name[:], b[i*16+12:i*16+15])
		for _, c := range s[i].Name {
			if c != ' ' && c != '*' && (c < 'A' || c > 'Z') {
				return s, fmt.Errorf("invalid initial")
			}
		}
		if b[i*16+15] != 0 {
			return s, fmt.Errorf("invalid record terminator")
		}
		if i > 0 && s[i].Digits.Uint64() > s[i-1].Digits.Uint64() {
			return s, fmt.Errorf("unordered high scores")
		}
	}
	return s, nil
}

// Store keeps writable records separate from installation assets; tests inject
// their own store and never consult the user's directory.
type Store interface {
	Load(table int) (Scores, error)
	Save(table int, s Scores) error
}
type FileStore struct{ Directory, SeedDirectory string }

func (f FileStore) Load(table int) (Scores, error) {
	return f.loadWithDefault(table, Defaults(table))
}

func (f FileStore) loadWithDefault(table int, fallback Scores) (Scores, error) {
	name := fmt.Sprintf("TABLE%d.HI", table)
	b, e := os.ReadFile(filepath.Join(f.Directory, name))
	if os.IsNotExist(e) {
		if f.SeedDirectory == "" {
			return fallback, nil
		}
		b, e = os.ReadFile(filepath.Join(f.SeedDirectory, name))
		if os.IsNotExist(e) {
			return fallback, nil
		}
	}
	if e != nil {
		return Scores{}, e
	}
	return DecodeScores(b)
}
func (f FileStore) Save(table int, s Scores) error {
	if e := os.MkdirAll(f.Directory, 0700); e != nil {
		return e
	}
	tmp, e := os.CreateTemp(f.Directory, ".high-*")
	if e != nil {
		return e
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, e = tmp.Write(s.MarshalBinary()); e == nil {
		e = tmp.Sync()
	}
	ce := tmp.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	return os.Rename(name, filepath.Join(f.Directory, fmt.Sprintf("TABLE%d.HI", table)))
}

// Installation defaults are bound at Load, never stored globally or by profile.
// Existing writable scores and optional .HI seeds keep their precedence.
type installationScoreStore struct {
	FileStore
	defaults [4]Scores
}

func (s installationScoreStore) Load(table int) (Scores, error) {
	return s.FileStore.loadWithDefault(table, s.defaults[table-1])
}
