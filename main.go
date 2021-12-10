package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Jille/convreq"
	"github.com/bertha/streisand/diskstore"
)

var (
	port      = flag.Int("port", 8080, "HTTP port to serve on")
	dataDir   = flag.String("datadir", "", "Path to data directory")
	cacheDir  = flag.String("cachedir", "", "Path to the cache directory")
	withFsync = flag.Bool("with-fsync", false,
		"Whether to fsync newly written blobs")
	peers Peers
)

func setAdditionalFlags() {
	flag.Var(&peers, "peer", "Provide url of peer")
}

type Peers []*url.URL

func (p *Peers) Set(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	*p = append(*p, u)
	return nil
}

func (p *Peers) String() string {
	strs := make([]string, 0, len(*p))

	for _, u := range *p {
		strs = append(strs, u.String())
	}
	return strings.Join(strs, " ")
}

var store *diskstore.Store
var xors *XorStore

func main() {
	setAdditionalFlags()
	flag.Parse()

	store = &diskstore.Store{
		Path:          *dataDir,
		BitsPerFolder: []uint8{8, 8},
		Fsync:         *withFsync,
	}
	store.Initialize()

	xors = &XorStore{
		LayerCount: 6,
		LayerDepth: 4,
		Path:       *cacheDir,
	}
	if err := xors.Initialize(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/blob/", convreq.Wrap(func(r *http.Request) convreq.HttpResponse {
		return handleGetBlob(r, true)
	}))
	http.HandleFunc("/internal/blob/", convreq.Wrap(func(r *http.Request) convreq.HttpResponse {
		return handleGetBlob(r, false)
	}))
	http.HandleFunc("/upload", convreq.Wrap(handlePostBlob))
	http.HandleFunc("/internal/upload", convreq.Wrap(handleInternalPostBlob))
	http.HandleFunc("/list", convreq.Wrap(handleGetList))
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
	})

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	log.Fatal(err)
}
