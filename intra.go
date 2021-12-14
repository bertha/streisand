package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// pushBlob sends a blob to a target server (given as host:port)
// Passing in an fh is optional, but if you do, it will be closed before returning.
func (s *server) pushBlob(target string, hash []byte, fh *os.File) error {
	// TODO: locking
	if fh == nil {
		var err error
		fh, err = s.store.Get(hash)
		if err != nil {
			return err
		}
	}
	defer fh.Close()
	st, err := fh.Stat()
	if err != nil {
		return err
	}
	u := url.URL{
		Scheme: "http",
		Host:   target,
		Path:   "/internal/upload",
	}
	req, err := http.NewRequest("POST", u.String(), fh)
	if err != nil {
		return err
	}
	req.Header.Set("Expect", "100-continue")
	req.Header.Set("X-StreiSANd-Hash", hex.EncodeToString(hash))
	req.Header.Set("Content-Length", fmt.Sprint(st.Size()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}
	return nil
}

func (s *server) pullBlob(target string, hash []byte) error {
	u := url.URL{
		Scheme: "http",
		Host:   target,
		Path:   "/internal/blob/" + hex.EncodeToString(hash),
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}
	w, err := s.store.NewWriter()
	if err != nil {
		return err
	}
	defer w.Abort()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := w.Close(); err != nil {
		return err
	}
	if bytes.Compare(hash, w.Hash()) != 0 {
		return fmt.Errorf("remote returned incorrect file: want %q, got %q", hex.EncodeToString(hash), hex.EncodeToString(w.Hash()))
	}
	if w.IsNew() {
		s.xors.Add((*Hash)(hash))
	}
	return nil
}
