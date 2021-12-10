package main

import (
	"fmt"
	"github.com/Jille/errchain"
	"log"
	"os"
	"path"
	"syscall"
)

type XorStore struct {
	LayerCount int
	LayerDepth int
	Path       string

	layers []Layer
}

func (s *XorStore) Initialize() (err error) {
	s.layers = make([]Layer, s.LayerCount)
	for i := range s.layers {

		layerName := fmt.Sprintf("xors-%d-layer-%d", s.LayerDepth, i)
		s.layers[i] = Layer{
			Path:     path.Join(s.Path, layerName),
			XorCount: 1 << ((i + 1) * s.LayerDepth),
		}
		err = s.layers[i].Initialize()
		if err != nil {
			return fmt.Errorf("init layer %d: %w", i, err)
		}
	}

	return
}

func (s *XorStore) Close() (err error) {
	for _, layer := range s.layers {
		errchain.Call(&err, layer.Close)
	}
	return
}

type Layer struct {
	Path     string
	XorCount int // number of 32-byte hashes

	file *os.File
	mmap []byte
}

func (l *Layer) Initialize() (err error) {
	// open or create the file that holds the xors of this layer
	l.file, err = os.OpenFile(l.Path,
		os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// check the file has the correct size, or, when it has size 0,
	// truncate the file to the correct size.
	expectedSize := pagesizemult(l.XorCount * BytesPerHash)

	fi, err := l.file.Stat()
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

		if err := l.file.Truncate(int64(expectedSize)); err != nil {
			return err
		}

		// check truncate worked
		fi, err = l.file.Stat()
		if err != nil {
			return err
		}
		if fi.Size() != int64(expectedSize) {
			return fmt.Errorf("failed to correct size of %s from "+
				"0 to %d", l.Path, expectedSize)
		}
	}

	l.mmap, err = syscall.Mmap(int(l.file.Fd()), 0, expectedSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED, // carry changes to the underlying file
	)
	if err != nil {
		return fmt.Errorf("mmap: %w", err)
	}

	return
}

func (l *Layer) Close() (err error) {
	errchain.Call(&err, func() error {
		return syscall.Munmap(l.mmap)
	})
	errchain.Call(&err, l.file.Close)

	return
}

var pagesize = os.Getpagesize()

func pagesizemult(size int) int {
	pagecount := size / pagesize
	if size%pagesize > 0 {
		pagecount++
	}
	return pagecount * pagesize
}
