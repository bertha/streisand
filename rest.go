package main

import (
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
)

type Hash [32]byte

func (h *Hash) UnmarshalText(text []byte) error {
    return nil
}

func (h *Hash) MarshalText() ([]byte, error) {
    return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) String() {
    hex.EncodeToString(h[:])
}

type Prefix struct {
	Hash   Hash
	Length int
}

func httpPathToHash(path string) (Hash, bool) {
	h, err := hex.DecodeString(strings.TrimPrefix(path, "/blob/"))
	if err != nil {
		return Hash{}, false
	}
	return Hash(h), true
}

func handleGetBlob(r *http.Request) convreq.HttpResponse {
	hash, ok := httpPathToHash(r.URL.Path)
	if !ok {
		return respond.BadRequest("invalid hash")
	}
	fh, err := store.Get(hash[:])
	if os.IsNotExist(err) {
		// TODO try with another
		//
		return respond.NotFound("blob not found")
	}
}
