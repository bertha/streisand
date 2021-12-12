package main

import (
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
)

func handlePostBlob(r *http.Request) convreq.HttpResponse {
	if r.Method != "POST" {
		return respond.MethodNotAllowed("Method Not Allowed")
	}
	// TODO: Simultaneously sync it to peers.

	hash, err := Post(r.Body)
	if err != nil {
		return respond.Error(err)
	}
	return respond.String(hex.EncodeToString(hash))
}

func handleInternalPostBlob(r *http.Request) convreq.HttpResponse {
	if r.Method != "POST" {
		return respond.MethodNotAllowed("Method Not Allowed")
	}
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
		go warnOnErr(checkXorsumOf(&h),
			"checking leaf xorsum of %s", &h)

		// TODO: See if there's a better code than HTTP 409 Conflict.
		return respond.OverrideResponseCode(respond.String("already exists"), 409)
	}

	hash, err := Post(r.Body)
	if err != nil {
		return respond.Error(err)
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

func handleDebugAddXor(r *http.Request) convreq.HttpResponse {
	if r.Method != "POST" {
		return respond.MethodNotAllowed("Method Not Allowed")
	}
	var h Hash
	n, err := hex.Decode(h[:], []byte(r.Header.Get("X-StreiSANd-Hash")))
	if err != nil {
		return respond.BadRequest("couldn't decode hash in X-StreiSANd-Hash")
	}
	if len(h) != n {
		return respond.BadRequest("wrong hash length in X-StreiSANd-Hash")
	}
	xors.Add(&h)
	xorsum := xors.GetLeaf(&h)
	return respond.String(xorsum.String())
}

func checkXorsumOf(h *Hash) (err error) {
	if *debug {
		log.Printf("checking xorsum of %s", h)
	}

	Lock()
	defer Unlock()

	// compute the difference between xorsum stored
	// and the xorsum computed from the disk store
	storedXorsum := xors.GetLeaf_AlreadyLocked(h)
	var computedXorsum Hash

	if err := store.Scan((*h)[:], (uint8)(xors.Depth()),
		func(hash []byte) {
			(*Hash)(hash).XorInto(computedXorsum[:])
		}); err != nil {

		return err
	}

	diff := storedXorsum.Xor(&computedXorsum)

	if diff.IsZero() {
		// all ok
		return
	}

	if diff.Equals(h) {
		log.Printf("warning: adding missing hash %s to xorsum", h)
		xors.Add_AlreadyLocked(h)
		return
	}

	// something serious is wrong;
	panic("corrupted xorsum table")

	// TODO: start repairing instead

	return
}
