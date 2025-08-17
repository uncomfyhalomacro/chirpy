package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/uncomfyhalomacro/chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

type postDataShape struct {
	UserID uuid.UUID `json:"user_id"`
	Body   string    `json:"body"`
}
type returnErrChirp struct {
	Err string `json:"error"`
}
type returnValidChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
type UserDetail struct {
	Email string `json:"email"`
}
type returnUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
		log.Println("Hit!")
		// your logic (e.g., increment the counter)
		// then hand off to next handler
	})
}

func (cfg *apiConfig) chirps(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		cfg.postChirps(w, r)
		return
	}
	if r.Method == "GET" {
		cfg.getChirps(w, r)
		return
	}
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	pathValue := r.PathValue("chirpID")
	if pathValue != "" {
		id, err := uuid.Parse(pathValue)
		if err != nil {
			msg := fmt.Sprintf("500 - %s", err)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 500)
			return
		}
		chirp, err := cfg.db.GetChirp(r.Context(), id)
		if err != nil {
			msg := fmt.Sprintf("404 - %s", err)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 404)
			return
		}
		respBody := returnValidChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		dat, errMarshal := json.Marshal(respBody)
		if errMarshal != nil {
			msg := fmt.Sprintf("500 - %s", errMarshal)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)
		return

	}
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Printf("%s\n", msg)
		http.Error(w, msg, 500)
		return
	}
	var chirpJSON []returnValidChirp
	for _, chirp := range chirps {
		chirpJSON = append(chirpJSON, returnValidChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	dat, errMarshal := json.Marshal(chirpJSON)
	if errMarshal != nil {
		msg := fmt.Sprintf("500 - %s", errMarshal)
		log.Printf("%s\n", msg)
		http.Error(w, msg, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) postChirps(w http.ResponseWriter, r *http.Request) {
	var postData postDataShape
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&postData)
	if err != nil {
		respBody := returnErrChirp{
			Err: fmt.Sprintf("%v", err),
		}
		dat, errMarshal := json.Marshal(respBody)
		if errMarshal != nil {
			msg := fmt.Sprintf("500 - %s", errMarshal)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	log.Printf("Before cleaned: %v\n", postData.Body)
	cleanedBody := cleanProfaneBody(postData.Body)
	log.Printf("After cleaned: %v\n", cleanedBody)
	if len(postData.Body) > 140 {
		respBody := returnErrChirp{
			Err: "Chirp is too long",
		}
		dat, errMarshal := json.Marshal(respBody)
		if errMarshal != nil {
			msg := fmt.Sprintf("500 - %s", errMarshal)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}
	params := database.CreateChirpParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      cleanedBody,
		UserID:    postData.UserID,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), params)
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Printf("failed to create chirp! %s\n", msg)
		http.Error(w, msg, 500)
		return
	}
	respBody := returnValidChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	dat, errMarshal := json.Marshal(respBody)
	if errMarshal != nil {
		msg := fmt.Sprintf("500 - %s", errMarshal)
		log.Printf("%s\n", msg)
		http.Error(w, msg, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func cleanProfaneBody(s string) string {
	fields := strings.Split(s, " ")
	badwords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	for idx, element := range fields {
		if badwords[strings.ToLower(element)] {
			fields[idx] = "****"
		}
	}
	return strings.Join(fields, " ")
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	n, err := w.Write([]byte(fmt.Sprintf(`<html>
<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>
</html>`, cfg.fileserverHits.Load())))
	if err != nil {
		log.Fatalln("Unable to write to response writer!")
	}
	log.Printf("Have written a total of %d bytes\n", n)
	log.Printf("Total hits: %d\n", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if dev := os.Getenv("PLATFORM"); dev != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
		return
	}
	// n, err := w.Write([]byte("OK"))
	// if err != nil {
	// 	log.Fatalln("Unable to write to response writer!")
	// }
	// log.Printf("Have written a total of %d bytes\n", n)
	// cfg.fileserverHits.Swap(0)
	// log.Println("Reset number of hits to 0")
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Database query error: %v", err)))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var postData UserDetail
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&postData)
	params := database.CreateUserParams{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     postData.Email,
	}
	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Printf("failed to create user! %s\n", msg)
		http.Error(w, msg, 500)
		return
	}

	responseJson := returnUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	dat, err := json.Marshal(responseJson)
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Println(msg)
		http.Error(w, msg, 500)
		return
	}
	w.WriteHeader(201)
	w.Write(dat)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v\n", dbURL, err)
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		db: dbQueries,
	}
	curdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v\n", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(curdir)))))
	mux.Handle("POST /api/chirps", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.chirps)))
	mux.Handle("GET /api/chirps", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.chirps)))
	mux.Handle("GET /api/chirps/{chirpID}", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.chirps)))
	mux.Handle("GET /api/healthz", apiCfg.middlewareMetricsInc(http.HandlerFunc(readiness)))
	mux.Handle("POST /api/users", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.createUser)))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.numberOfHits))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.reset))
	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = mux
	if err = server.ListenAndServe(); err != nil {
		log.Fatalf("listen and serve failed: %v\n", err)
	}
	if err = server.Close(); err != nil {
		log.Fatalf("closing the server failed: %v\n", err)
	}

}
