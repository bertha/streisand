package main

import (
	"net/http"
    "encoding/hex"
)

type Hash [32]byte

func (h *Hash) String() {
    hex.EncodeToString(
}

type Prefix struct {
    Hash Hash
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
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handlePostBlob(w, r)
		} else if r.Method == "GET" {
			handleGetBlob(w, r)
		}
	})
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
    }

	http.ListenAndServe(":8080", nil)
}
