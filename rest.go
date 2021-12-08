package main

import (
	"encoding/hex"
	"fmt"
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
	var ret Hash
	n, err := hex.Decode(ret[:], []byte(strings.TrimPrefix(path, "/blob/")))
	if err != nil {
		return Hash{}, false
	}
	if len(ret) != n {
		return Hash{}, false
	}
	return ret, true
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
	st, err := fh.Stat()
	if err != nil {
		return respond.Error(err)
	}
	hdrs := http.Header{}
	hdrs.Set("Content-Length", fmt.Sprint(st.Size()))
	return respond.WithHeaders(respond.Reader(fh), hdrs)
}

func handleGetList(r *http.Request) convreq.HttpResponse {
	var ret []string
	if err := store.Scan([]byte("1d"), 0, func(h []byte) {
		ret = append(ret, hex.EncodeToString(h))
	}); err != nil {
		return respond.Error(err)
	}
	return respond.String(fmt.Sprintf("%d entries\n", len(ret)) + strings.Join(ret, "\n"))
}
