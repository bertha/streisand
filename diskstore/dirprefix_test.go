package diskstore

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPathinizer(t *testing.T) {
	tests := []struct {
		bitsPerFolder []uint8
		input         []byte
		want          []string
		wantFullPath  string
	}{
		{
			bitsPerFolder: []uint8{8, 8},
			input:         []byte{0xc0, 0xde, 0xff},
			want:          []string{"c0", "de"},
			wantFullPath:  "/tmp/blobs/c0/de/c0deff",
		},
		{
			bitsPerFolder: []uint8{8, 4},
			input:         []byte{0xc0, 0xde, 0xff},
			want:          []string{"c0", "d"},
			wantFullPath:  "/tmp/blobs/c0/d/c0deff",
		},
		{
			bitsPerFolder: []uint8{8, 8, 2},
			input:         []byte{0xc0, 0xde, 0xff},
			want:          []string{"c0", "de", "3"},
			wantFullPath:  "/tmp/blobs/c0/de/3/c0deff",
		},
	}
	for _, tc := range tests {
		c := Store{
			Path:          "/tmp/blobs",
			BitsPerFolder: tc.bitsPerFolder,
		}
		c.Initialize()

		if diff := cmp.Diff(tc.want, c.DirsFor(tc.input)); diff != "" {
			t.Errorf("DirsFor(%v, %v) failed: %v", tc.bitsPerFolder, tc.input, diff)
		}
		if diff := cmp.Diff(tc.wantFullPath, c.FullPath(tc.input)); diff != "" {
			t.Errorf("FullPath(%v, %v) failed: %v", tc.bitsPerFolder, tc.input, diff)
		}
	}
}
