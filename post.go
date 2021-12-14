package main

import (
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/Jille/convreq"
	"github.com/Jille/convreq/respond"
)

func (s *server) handlePostBlob(r *http.Request) convreq.HttpResponse {
	if r.Method != "POST" {
		return respond.MethodNotAllowed("Method Not Allowed")
	}
	// TODO: Simultaneously sync it to peers.

	hash, err := s.Post(r.Body)
	if err != nil {
		return respond.Error(err)
	}
	return respond.String(hex.EncodeToString(hash))
}

func (s *server) handleInternalPostBlob(r *http.Request) convreq.HttpResponse {
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

	has, err := s.store.Has(h[:])
	if err != nil {
		return respond.Error(err)
	}
	if has {
		go func() {
			warnOnErr(s.checkXorsumOf(&h),
				"checking leaf xorsum of %s", &h)
		}()

		// TODO: See if there's a better code than HTTP 409 Conflict.
		return respond.OverrideResponseCode(respond.String("already exists"), 409)
	}

	// TODO: abort if hash is not equal to h (X-StreiSANd-Hash)
	hash, err := s.Post(r.Body)
	if err != nil {
		return respond.Error(err)
	}
	return respond.String(hex.EncodeToString(hash))
}

func (s *server) Post(blob io.ReadCloser) (hash []byte, err error) {
	w, err := s.store.NewWriter()
	if err != nil {
		return
	}
	defer w.Abort()

	if _, err = io.Copy(w, blob); err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err = blob.Close(); err != nil {
		return
	}

	if err = w.Close(); err != nil {
		return
	}

	hash = w.Hash()

	if w.IsNew() {
		s.xors.Add((*Hash)(hash))
	}

	return
}

func (s *server) handleDebugAddXor(r *http.Request) convreq.HttpResponse {
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.xors.Add(&h)
	xorsum := s.xors.GetLeaf(&h)
	return respond.String(xorsum.String())
}

func (s *server) checkXorsumOf(h *Hash) (err error) {
	if s.conf.Debug {
		log.Printf("checking xorsum of %s", h)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// compute the difference between xorsum stored
	// and the xorsum computed from the disk store
	storedXorsum := s.xors.GetLeaf(h)
	var computedXorsum Hash

	if err := s.store.Scan(h[:], uint8(s.xors.Depth()),
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
		s.xors.Add(h)
		return
	}

	// something serious is wrong;
	panic("corrupted xorsum table")

	// TODO: start repairing instead

	return
}
