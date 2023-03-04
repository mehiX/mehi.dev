package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	if err := http.ListenAndServe("127.0.0.1:9090", nil); err != nil {
		log.Fatal(err)
	}
}
