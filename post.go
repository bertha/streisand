package main

import (
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
	"github.com/bertha/streisand/diskstore"
)

var store *diskstore.Store

func handlePostBlob(r *http.Request) convreq.HttpResponse {
	// TODO handle the case of given hash

	w, err := store.NewWriter()
	if err != nil {
		return respond.Error(err)
	}
	defer w.Abort()
	if _, err := io.Copy(w, r.Body); err != nil {
		return respond.Error(err)
	}
	if err := r.Body.Close(); err != nil {
		return respond.Error(err)
	}
	if err := w.Close(); err != nil {
		return respond.Error(err)
	}
	return respond.String(hex.EncodeToString(w.Hash()))
}
