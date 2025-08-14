package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
		// your logic (e.g., increment the counter)
		// then hand off to next handler
	})
}

func readiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	n, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatalln("Unable to write to response writer!")
	}
	log.Printf("Have written a total of %d bytes\n", n)
}

func (cfg *apiConfig) numberOfHits(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	n, err := w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
	if err != nil {
		log.Fatalln("Unable to write to response writer!")
	}
	log.Printf("Have written a total of %d bytes\n", n)
}

func (cfg *apiConfig) reset(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	n, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatalln("Unable to write to response writer!")
	}
	log.Printf("Have written a total of %d bytes\n", n)
	cfg.fileserverHits.Swap(0)
	log.Println("Reset number of hits to 0")
}

func main() {
	apiCfg := apiConfig{}
	curdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v\n", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(curdir)))))
	mux.Handle("GET /healthz", apiCfg.middlewareMetricsInc(http.HandlerFunc(readiness)))
	mux.Handle("GET /metrics", http.HandlerFunc(apiCfg.numberOfHits))
	mux.Handle("POST /reset", http.HandlerFunc(apiCfg.reset))
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
