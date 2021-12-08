package main

import (
	"flag"
	"net/http"

	"github.com/Jille/convreq"
	"github.com/bertha/streisand/diskstore"
)

var (
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

	// TODO lees config
	http.HandleFunc("/blob/", convreq.Wrap(handleGetBlob))
	http.HandleFunc("/upload", convreq.Wrap(handlePostBlob))
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
	})

	http.ListenAndServe(":8080", nil)
}
