package main

////////////////////////////////////////////////////////////////////////////////

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/sabhiram/pngr"
)

////////////////////////////////////////////////////////////////////////////////

var (
	inputFile string
)

////////////////////////////////////////////////////////////////////////////////

func fatalOnError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	bs, err := ioutil.ReadFile(inputFile)
	fatalOnError(err)

	r1, err := pngr.NewReader(bs, nil)
	fatalOnError(err)

	r2, err := pngr.NewReader(bs, &pngr.ReaderOptions{
		IncludedChunkTypes: []string{"tEXt"},
	})
	fatalOnError(err)

	for i, r := range []*pngr.Reader{r1, r2} {
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

func init() {
	flag.StringVar(&inputFile, "input", "", "input file name for the png")
	flag.Parse()
	if len(inputFile) == 0 {
		fmt.Printf("Fatal error: No input file specified!\n")
		os.Exit(1)
	}
}
