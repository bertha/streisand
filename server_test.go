package streisand

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	s, err := NewServer(ServerConfig{
		DataDir:   t.TempDir(),
		CacheDir:  t.TempDir(),
		WithFsync: true,
		GetPeers:  nil,
	})

	if err != nil {
		t.Fatal(err)
	}

	// upload file
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("POST", "/upload",
		strings.NewReader("test")))
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}

	hash, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(hash, []byte("9f86d081884c7d659a2feaa0c55"+
		"ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")) != 0 {
		t.Fatal("upload returned invalid hash")
	}

	// download it
	w = httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET",
		"/blob/"+string(hash),
		strings.NewReader("test")))
	resp = w.Result()

	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}

	blob, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(blob, []byte("test")) != 0 {
		t.Fatal("server didn't return uploaded file")
	}

	err = s.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// TODO: add benchmarks for simultaneous up-/downloading
