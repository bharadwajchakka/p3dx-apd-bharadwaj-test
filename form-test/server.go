package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)
	addr := ":8080"
	fmt.Printf("Serving web UI at http://localhost%s/\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
