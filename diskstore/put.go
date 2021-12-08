package diskstore

import (
	"crypto/sha256"
	"hash"
	"io"
	"os"
	"path/filepath"
)

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
