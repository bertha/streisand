package main

import (
	"encoding/hex"
	"net/http"
	"os"

	"github.com/Jille/convreq"
)

type Hash [32]byte

func (h *Hash) UnmarshalText(text []byte) error {
    if len(text) != 32 {
        return
    }
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

func got(hash Hash) bool {
}
func httpPathToHash(path string) Hash {
}
func open(hash Hash) (os.File, error) {
}

func handleGetBlob(w http.ResponseWriter, r *http.Request) {
	hash, err := httpPathToHash(r.Path)
	if err != nil {
		http.Error(w, err.String(), 400)
		return
	}
	fh, err := open(hash)
	if err == ErrNotFound {
		// TODO try with another
		//
		return
	}
	// TODO serve back file
}

func main() {
	// TODO lees config
	http.HandleFunc("/blob/", func(w http.ResponseWriter, r *http.Request) {
		handlePostBlob(w, r)
	})
	http.HandleFunc("/upload", convreq.Wrap(handlePostBlob))
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
    })

	http.ListenAndServe(":8080", nil)
}
