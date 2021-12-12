package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
)

const BytesPerHash = 32

type Hash [BytesPerHash]byte

func (h *Hash) UnmarshalText(text []byte) error {
	return nil
}

func (h *Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h *Hash) XorInto(s []byte) {
	if len(s) != BytesPerHash {
		panic("xoring hash into slice of incorrect size")
	}
	// TODO:  optimize?
	for i := 0; i < BytesPerHash; i++ {
		s[i] ^= h[i]
	}
}

func (h *Hash) Xor(other *Hash) Hash {
	result := *other
	h.XorInto(result[:])
	return result
}

func (h *Hash) Equals(other *Hash) bool {
	return bytes.Compare((*h)[:], (*other)[:]) == 0
}

var zeroHash Hash

func (h *Hash) IsZero() bool {
	return h.Equals(&zeroHash)
}

func (h *Hash) PrefixToNumber(prefixLength uint) uint32 {
	p32 := binary.BigEndian.Uint32(h[0:4])
	return p32 >> (32 - prefixLength)
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

func handleGetBlob(r *http.Request, allowForward bool) convreq.HttpResponse {
	if r.Method != "GET" {
		return respond.MethodNotAllowed("Method Not Allowed")
	}
	hash, ok := httpPathToHash(r.URL.Path)
	if !ok {
		return respond.BadRequest("invalid hash")
	}
	fh, err := store.Get(hash[:])
	if os.IsNotExist(err) {
		if allowForward {
			// TODO try with another
		}
		return respond.NotFound("blob not found")
	}
	st, err := fh.Stat()
	if err != nil {
		return respond.Error(err)
	}
	hdrs := http.Header{}
	hdrs.Set("Content-Length", fmt.Sprint(st.Size()))
	hdrs.Set("Last-Modified", st.ModTime().UTC().Format(http.TimeFormat))
	hdrs.Set("Etag", fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:])))
	hdrs.Set("Cache-Control", "max-age=604800, immutable, stale-if-error=604800")
	return respond.WithHeaders(respond.Reader(fh), hdrs)
}

func handleGetList(r *http.Request) convreq.HttpResponse {
	if r.Method != "GET" {
		return respond.MethodNotAllowed("Method Not Allowed")
	}
	var ret []string
	if err := store.Scan([]byte("1d"), 0, func(h []byte) {
		ret = append(ret, hex.EncodeToString(h))
	}); err != nil {
		return respond.Error(err)
	}
	return respond.String(fmt.Sprintf("%d entries\n", len(ret)) + strings.Join(ret, "\n"))
}

func warnOnErr(err error, message string, v ...interface{}) {
	if err == nil {
		return
	}
	log.Printf("warning: %s: %v", fmt.Sprintf(message, v...), err)
}
