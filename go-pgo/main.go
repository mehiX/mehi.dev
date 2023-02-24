package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"gitlab.com/golang-commonmark/markdown"

	_ "net/http/pprof"
)

func render(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	src, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	md := markdown.New(
		markdown.XHTMLOutput(true),
		markdown.Typographer(true),
		markdown.Linkify(true),
		markdown.Tables(true),
	)

	var buf bytes.Buffer
	if err := md.Render(&buf, src); err != nil {
		log.Printf("error converting markdown: %v", err)
		http.Error(w, "malformed markdown", http.StatusBadRequest)
		return
	}

	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func main() {
	http.HandleFunc("/render", render)
	addr := "127.0.0.1:8080"
	log.Printf("Serving on %s...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
