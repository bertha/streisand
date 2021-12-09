package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Jille/convreq"
	"github.com/bertha/streisand/diskstore"
)

var (
	port      = flag.Int("port", 8080, "HTTP port to serve on")
	dataDir   = flag.String("datadir", "", "Path to datadir")
	withFsync = flag.Bool("with-fsync", false, "Whether to fsync newly written blobs")
)

var store *diskstore.Store

func main() {
	flag.Parse()

	store = &diskstore.Store{
		Path:          *dataDir,
		BitsPerFolder: []uint8{8, 8},
		Fsync:         *withFsync,
	}
	store.Initialize()

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
