package pngr

////////////////////////////////////////////////////////////////////////////////

import (
	"io"
	"io/ioutil"
	"testing"
)

////////////////////////////////////////////////////////////////////////////////

const (
	redPng = "./fixtures/red.png"
)

////////////////////////////////////////////////////////////////////////////////

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Fatal error: %s\n", err.Error())
	}
}

func TestReader(t *testing.T) {
	bs, err := ioutil.ReadFile(redPng)
	fatalIfError(t, err)

	for _, tc := range []struct {
		data     []byte
		filter   []string
		isErr    bool
		isBadCrc bool
		segments int
	}{
		// Negative test cases.
		{data: []byte{1, 2, 3}, filter: []string{`IHDR`}, isErr: true, segments: 0},
		{data: []byte{1, 2, 3, 4, 5, 6, 7, 8}, filter: []string{`IHDR`}, isErr: true, segments: 0},

		// Corrupt chunks.
		{data: append(pngMagic, []byte{1, 2, 3}...), filter: []string{`IHDR`}, isErr: false, segments: 0},
		{data: append(pngMagic, []byte{1, 2, 3, 4, 5}...), filter: []string{`IHDR`}, isErr: false, segments: 0},
		{data: append(pngMagic, []byte{0, 0, 0, 0, 'I', 'E', 'N', 'D', 1}...), filter: []string{`IHDR`}, isErr: false, segments: 0},
		{data: append(pngMagic, []byte{0, 0, 0, 0, 'I', 'E', 'N', 'D', 1, 2, 3, 4}...), filter: []string{`IHDR`}, isErr: false, isBadCrc: true, segments: 0},

		// Positive test cases.
		{data: bs, filter: []string{`IHDR`}, isErr: false, segments: 1},
		{data: bs, filter: []string{`IDAT`}, isErr: false, segments: 1},
		{data: bs, filter: []string{`IEND`}, isErr: false, segments: 1},
		{data: bs, filter: nil, isErr: false, segments: 3},
	} {
		var opts *ReaderOptions
		if tc.filter != nil {
			opts = &ReaderOptions{
				IncludedChunkTypes: tc.filter,
			}
		}
		r, err := NewReader(tc.data, opts)

		if tc.isErr == true {
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
			continue
		}
		fatalIfError(t, err)

		ct := 0
		for {
			_, err := r.Next()
			if err == io.EOF {
				break
			}
			if tc.isBadCrc {
				if err != ErrBadCRC {
					t.Errorf("Expected bad crc, got other error!\n")
				} else {
					break
				}
			} else {
				fatalIfError(t, err)
				ct += 1
			}
		}

		if ct != tc.segments {
			t.Errorf("Expected %d segments, got %d\n", tc.segments, ct)
		}
	}
}
