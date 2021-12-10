package main

import (
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
)

func handlePostBlob(r *http.Request) convreq.HttpResponse {
	// TODO: Simultaneously sync it to peers.

	hash, err := Post(r.Body)
	if err != nil {
		respond.Error(err)
	}
	return respond.String(hex.EncodeToString(hash))
}

func handleInternalPostBlob(r *http.Request) convreq.HttpResponse {
	var h Hash
	n, err := hex.Decode(h[:], []byte(r.Header.Get("X-StreiSANd-Hash")))
	if err != nil {
		return respond.BadRequest("couldn't decode hash in X-StreiSANd-Hash")
	}
	if len(h) != n {
		return respond.BadRequest("wrong hash length in X-StreiSANd-Hash")
	}
	has, err := store.Has(h[:])
	if err != nil {
		return respond.Error(err)
	}
	if has {
		// TODO: See if there's a better code than HTTP 409 Conflict.
		return respond.OverrideResponseCode(respond.String("already exists"), 409)
	}

	hash, err := Post(r.Body)
	if err != nil {
		respond.Error(err)
	}
	return respond.String(hex.EncodeToString(hash))
}

func Post(blob io.ReadCloser) (hash []byte, err error) {
	w, err := store.NewWriter()
	if err != nil {
		return
	}
	defer w.Abort()

	if _, err = io.Copy(w, blob); err != nil {
		return
	}

	if err = blob.Close(); err != nil {
		return
	}
	if err = w.Close(); err != nil {
		return
	}

	hash = w.Hash()

	if w.IsNew() {
		xors.Add((*Hash)(hash))
	}

	return
}
