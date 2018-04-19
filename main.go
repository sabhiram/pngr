package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
)

var (
	pngMagic = []byte{137, 80, 78, 71, 13, 10, 26, 10}
)

type Reader struct {
	buf  *bytes.Buffer
	opts *ReaderOptions
}

type Chunk struct {
	// ----------------------------------------------------------------
	// |  Length    |  Chunk Type |       ... Data ...       |  CRC   |
	// ----------------------------------------------------------------
	//    4 bytes       4 bytes         `Length` bytes         4 bytes
	Length    uint32
	ChunkType string
	Data      []byte
	Crc       uint32
}

type ReaderOptions struct {
	IncludedChunkTypes []string
}

func NewReader(data []byte, opts *ReaderOptions) (*Reader, error) {
	n := len(pngMagic)
	buf := bytes.NewBuffer(data)
	magic := buf.Next(n)

	if len(magic) != n {
		return nil, errors.New("missing png file header")
	}

	i := 0
	for ; i < n; i++ {
		if magic[i] != pngMagic[i] {
			break
		}
	}
	if i != n {
		return nil, errors.New("missing png file header")
	}

	return &Reader{
		buf:  buf,
		opts: opts,
	}, nil
}

func (r *Reader) includesChunkType(ct string) bool {
	if r.opts == nil {
		return true
	}
	for _, v := range r.opts.IncludedChunkTypes {
		if v == ct {
			return true
		}
	}
	return false
}

func (r *Reader) Next() (*Chunk, error) {
	chunk := &Chunk{}

	for r.buf.Len() >= 12 {
		// V
		// ----------------------------------------------------------------
		// |  Length    |  Chunk Type |       ... Data ...       |  CRC   |
		// ----------------------------------------------------------------
		//    4 bytes       4 bytes         `Length` bytes         4 bytes

		err := binary.Read(r.buf, binary.BigEndian, &chunk.Length)
		if err != nil {
			break
		}

		//        buf = V
		// ----------------------------------------------------------------
		// |  Length    |  Chunk Type |       ... Data ...       |  CRC   |
		// ----------------------------------------------------------------
		//    4 bytes       4 bytes         `Length` bytes         4 bytes

		minLen := 4 + int(chunk.Length) + 4
		if r.buf.Len() < minLen {
			break
		}

		ctbs := r.buf.Next(4)
		chunk.ChunkType = string(ctbs)
		chunk.Data = r.buf.Next(int(chunk.Length))

		err = binary.Read(r.buf, binary.BigEndian, &chunk.Crc)
		if err != nil {
			break
		}

		expCrc := crc32.ChecksumIEEE(append(ctbs, chunk.Data...))
		if expCrc != chunk.Crc {
			return nil, errors.New("bad crc for chunk")
		}

		if r.includesChunkType(chunk.ChunkType) {
			return chunk, nil
		}
	}

	return nil, io.EOF
}

func fatalOnError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	bs, err := ioutil.ReadFile("../png-embed/out.png")
	fatalOnError(err)

	r1, err := NewReader(bs, nil)
	fatalOnError(err)

	r2, err := NewReader(bs, &ReaderOptions{
		IncludedChunkTypes: []string{"tEXt"},
	})
	fatalOnError(err)

	for i, r := range []*Reader{r1, r2} {
		fmt.Printf("Using reader #%d\n", i+1)
		for {
			chunk, err := r.Next()
			if err == io.EOF {
				break
			} else {
				fatalOnError(err)
			}

			fmt.Printf("Got chunk: Type: %s Len: %d ", chunk.ChunkType, len(chunk.Data))
			if chunk.ChunkType == "tEXt" {
				fmt.Printf("Value: %s", chunk.Data)
			}
			fmt.Printf("\n")
		}
	}
}
