package diskstore

import (
	"crypto/rand"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRetrieval(t *testing.T) {
	c := Store{
		Path:          t.TempDir(),
		Fsync:         true,
		BitsPerFolder: []uint8{8, 8, 2},
	}
	c.Initialize()

	tests := []struct {
		genBytes int
	}{
		{
			genBytes: 2,
		},
		{
			genBytes: 2048,
		},
	}
	for _, tc := range tests {
		data := make([]byte, tc.genBytes)
		rand.Read(data)
		w, err := c.NewWriter()
		if err != nil {
			t.Fatalf("NewWriter() failed: %v", err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("Writer.Write() failed: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Writer.Close() failed: %v", err)
		}
		h := w.Hash()

		fh, err := c.Get(h)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		retr, err := ioutil.ReadAll(fh)
		if err != nil {
			t.Fatalf("Get.ReadAll() failed: %v", err)
		}
		if err := fh.Close(); err != nil {
			t.Fatalf("Get.Close() failed: %v", err)
		}

		if diff := cmp.Diff(data, retr); diff != "" {
			t.Fatalf("Retrieved data was different: %s", diff)
		}
	}
}
