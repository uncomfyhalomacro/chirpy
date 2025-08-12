package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = mux
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("listen and serve failed: %v\n", err)
	}
	if err := server.Close(); err != nil {
		log.Fatalf("closing the server failed: %v\n", err)
	}

}
