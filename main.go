package main

import (
	"net/http"
)

type Hash [32]byte

Hash.Xor(

type Prefix struct {
    Hash Hash
    Length int
}

type PrefixQuery struct {
    Prefix Prefix
    Depth uint
}

type Query struct {
    Got []Hash
    Prefixes []PrefixRequest
}

type QueryResponse struct {
    Got []bool
    Prefixes []map[int]string
}

func got(hash string) bool {
}
func validHash(hash string) bool {
}
func hashToPath(hash string) string {
}
func httpPathToHash(path string) {
}
func open(hash string) (os.File, error) {
}

func handlePostBlob(w http.ResponseWriter, r *http.Request) {
    // TODO handle the case of given hash

    // TODO get temp file
    
    // TODO move tempfile to hash
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
