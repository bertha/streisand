package diskstore

import (
	"bytes"
	"encoding/hex"
	"math"
	"path/filepath"

	"github.com/icza/bitio"
)

type Store struct {
	Fsync         bool
	Path          string
	BitsPerFolder []uint8

	minBytes int
}

func (s *Store) Initialize() {
	var minBits uint8
	for _, n := range s.BitsPerFolder {
		minBits += n
	}
	s.minBytes = int(math.Ceil(float64(minBits) / 8.0))
}

func (s Store) HasEnoughBits(hash []byte) bool {
	return len(hash) >= s.minBytes
}

const hextable = "0123456789abcdef"

func (s Store) DirsFor(hash []byte) []string {
	br := bitio.NewReader(bytes.NewReader(hash))
	dirs := make([]string, len(s.BitsPerFolder))
	for i, n := range s.BitsPerFolder {
		rem := int(n)
		d := make([]byte, 0, n/8+1)
		for rem > 0 {
			v, err := br.ReadBits(min(uint8(rem), 4))
			if err != nil {
				// This should never happen because the input reader shouldn't give errors.
				panic(err)
			}
			rem -= 4
			d = append(d, hextable[v])
		}
		dirs[i] = string(d)
	}
	return dirs
}

func min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

func (s Store) FullPath(hash []byte) string {
	parts := make([]string, 1, 4)
	parts[0] = s.Path
	parts = append(parts, s.DirsFor(hash)...)
	parts = append(parts, hex.EncodeToString(hash))
	return filepath.Join(parts...)
}
