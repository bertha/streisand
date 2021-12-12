package main

import (
	"fmt"
	"github.com/Jille/errchain"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

func (s *XorStore) Add(h *Hash) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Add_AlreadyLocked(h)
}

func (s *XorStore) Add_AlreadyLocked(h *Hash) {
	for i := range s.layers {
		s.layers[i].Add(h)
	}
}

func (s *XorStore) GetLeaf(h *Hash) Hash {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.GetLeaf_AlreadyLocked(h)
}

func (s *XorStore) GetLeaf_AlreadyLocked(h *Hash) Hash {
	return s.layers[len(s.layers)-1].Get(h)
}

func (s *XorStore) Lock() {
	s.mutex.Lock()
}

func (s *XorStore) Unlock() {
	s.mutex.Unlock()
}

type XorStore struct {
	LayerCount int
	LayerDepth int
	Path       string

	layers []Layer
	mutex  sync.RWMutex
}

func (s *XorStore) Depth() int {
	return s.LayerCount * s.LayerDepth
}

func (s *XorStore) Initialize() (err error) {
	s.layers = make([]Layer, s.LayerCount)
	for i := range s.layers {

		layerName := fmt.Sprintf("xors-%d-layer-%d", s.LayerDepth, i)
		s.layers[i] = Layer{
			Path:         filepath.Join(s.Path, layerName),
			PrefixLength: uint((i + 1) * s.LayerDepth),
		}
		err = s.layers[i].Initialize()
		if err != nil {
			return fmt.Errorf("init layer %d: %w", i, err)
		}
	}

	return
}

func (s *XorStore) Close() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, layer := range s.layers {
		errchain.Call(&err, layer.Close)
	}
	return
}

type Layer struct {
	Path         string
	PrefixLength uint

	mmap []byte
}

func (l *Layer) Initialize() (err error) {
	if l.PrefixLength > 32 {
		return fmt.Errorf("prefix length too great:  %d > 32",
			l.PrefixLength)
	}

	// open or create the file that holds the xors of this layer
	f, err := os.OpenFile(l.Path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// check the file has the correct size, or, when it has size 0,
	// truncate the file to the correct size.
	xorCount := 1 << l.PrefixLength
	expectedSize := pagesizemult(xorCount * BytesPerHash)

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if size := fi.Size(); size != int64(expectedSize) {
		if size != 0 {
			// might be a misconfiguration, so let's not
			// try to be clever
			return fmt.Errorf("something is wrong: "+
				"%s has size %d instead of %d",
				l.Path, size, expectedSize)
		}

		// file was probably just created, so let us fix its size
		log.Printf("truncating %v to %d", l.Path, expectedSize)

		if err := f.Truncate(int64(expectedSize)); err != nil {
			return err
		}

		// check truncate worked
		fi, err = f.Stat()
		if err != nil {
			return err
		}
		if fi.Size() != int64(expectedSize) {
			return fmt.Errorf("failed to correct size of %s from "+
				"0 to %d", l.Path, expectedSize)
		}
	}

	l.mmap, err = syscall.Mmap(int(f.Fd()), 0, expectedSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED, // carry changes to the underlying file
	)
	if err != nil {
		return fmt.Errorf("mmap: %w", err)
	}

	// the file needn't be kept open for the mmap to persist
	if err := f.Close(); err != nil {
		return err
	}

	return
}

func (l *Layer) Close() (err error) {
	return syscall.Munmap(l.mmap)
}

func (l *Layer) Add(h *Hash) {
	idx := BytesPerHash * h.PrefixToNumber(l.PrefixLength)
	h.XorInto(l.mmap[idx : idx+BytesPerHash])
}

func (l *Layer) Get(h *Hash) Hash {
	idx := BytesPerHash * h.PrefixToNumber(l.PrefixLength)
	return *(*Hash)(l.mmap[idx : idx+BytesPerHash])
}

var pagesize = os.Getpagesize()

func pagesizemult(size int) int {
	pagecount := size / pagesize
	if size%pagesize > 0 {
		pagecount++
	}
	return pagecount * pagesize
}
