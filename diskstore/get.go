package diskstore

import (
	"errors"
	"os"
)

func (s *Store) Get(hash []byte) (*os.File, error) {
	if !s.HasEnoughBits(hash) {
		return nil, errors.New("hash is too short")
	}
	return os.Open(s.FullPath(hash))
}
