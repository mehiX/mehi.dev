package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
)

var (
	source = flag.String("source", "", "path to markdown file to upload")
	addr   = flag.String("addr", "http://localhost:8080", "address of server")
)

func generateLoad(n int) error {

	if *source == "" {
		return fmt.Errorf("-source must be set to a markdown source file")
	}
	if *addr == "" {
		return fmt.Errorf("-addr must be set to the address of a server")
	}

	src, err := os.ReadFile(*source)
	if err != nil {
		return fmt.Errorf("error reading source: %v", err)
	}

	reader := bytes.NewReader(src)

	url := *addr + "/render"

	for i := 0; i < n; i++ {
		reader.Seek(0, io.SeekStart)

		resp, err := http.Post(url, "text/markdown", reader)
		if err != nil {
			return fmt.Errorf("error writing request: %v", err)
		}
		defer resp.Body.Close()

		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}
	}

	return nil
}

func main() {
	flag.Parse()

	if err := generateLoad(math.MaxInt); err != nil {
		log.Fatal(err)
	}
}
