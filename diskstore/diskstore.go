package diskstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"math"
	"os"
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

func (s *Store) NewWriter() (*Writer, error) {
	fh, err := os.CreateTemp(s.Path, "")
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	mw := io.MultiWriter(fh, h)
	return &Writer{
		s:            s,
		fh:           fh,
		hasher:       h,
		mw:           mw,
		needsClosing: true,
		needsRemoval: true,
	}, nil
}

type Writer struct {
	s      *Store
	fh     *os.File
	hasher hash.Hash
	mw     io.Writer
	result []byte

	needsClosing bool
	needsRemoval bool
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.mw.Write(b)
}

func (w *Writer) Close() error {
	defer w.Abort()
	sum := w.hasher.Sum(nil)
	fullPath := w.s.FullPath(sum)
	if _, err := os.Stat(fullPath); err == nil {
		// Already exists.
		w.result = sum
		return nil
	}
	dirs := w.s.DirsFor(sum)
	if w.s.Fsync {
		if err := w.fh.Sync(); err != nil {
			return err
		}
	}
	if err := w.fh.Close(); err != nil {
		return err
	}
	w.needsClosing = false
	if err := os.Rename(w.fh.Name(), fullPath); err != nil {
		// Create parent directories first.
		p := w.s.Path
		for _, d := range dirs {
			prevP := p
			p = filepath.Join(p, d)
			if err := os.Mkdir(p, 0777); err != nil {
				if os.IsExist(err) {
					continue
				}
				return err
			}
			if w.s.Fsync {
				if err := syncDir(prevP); err != nil {
					return err
				}
			}
		}

		if err := os.Rename(w.fh.Name(), fullPath); err != nil {
			return err
		}
	}

	w.needsRemoval = false
	if err := syncDir(filepath.Dir(fullPath)); err != nil {
		return err
	}
	w.result = sum
	return nil
}

func (w *Writer) Hash() []byte {
	if w.result == nil {
		panic("blobstorage.Writer.Hash() called without successful Close()")
	}
	return w.result
}

func (w *Writer) Abort() {
	if w.needsClosing {
		_ = w.fh.Close()
		w.needsClosing = false
	}
	if w.needsRemoval {
		_ = os.Remove(w.fh.Name())
		w.needsRemoval = true
	}
}

func syncDir(dirName string) error {
	dh, err := os.Open(dirName)
	if err != nil {
		return err
	}
	if err := dh.Sync(); err != nil {
		dh.Close()
		return err
	}
	return dh.Close()
}

func (s *Store) Get(hash []byte) (*os.File, error) {
	if !s.HasEnoughBits(hash) {
		return nil, errors.New("hash is too short")
	}
	return os.Open(s.FullPath(hash))
}