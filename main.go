package main

import (
	"database/sql"
	"sort"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/uncomfyhalomacro/chirpy/internal/auth"
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
	tokenSecret    string
	polkaSecret    string
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

type UserLoginDetail struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Expiry   int64  `json:"expires_in_seconds"`
}

type returnUser struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
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
	if r.Method == "DELETE" {
		cfg.deleteChirps(w, r)
		return
	}
}

func (cfg *apiConfig) deleteChirps(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("no bearer")
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("jwt error: %v", err)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
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
		if chirp.UserID != userID {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		w.WriteHeader(204)
		return

	}
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	sortKind := r.URL.Query().Get("sort")
	author_id := r.URL.Query().Get("author_id")
	pathValue := r.PathValue("chirpID")
	if author_id != "" && pathValue != "" {
		http.Error(w, http.StatusText(409), 409)
		return
	}
	if author_id != "" && pathValue == "" {
		id, err := uuid.Parse(author_id)
		if err != nil {
			msg := fmt.Sprintf("500 - %s", err)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 500)
			return
		}
		chirps, err := cfg.db.GetChirpsByUserID(r.Context(), id)
		var returnValidChirps []returnValidChirp
		if err != nil {
			msg := fmt.Sprintf("404 - %s", err)
			log.Printf("%s\n", msg)
			http.Error(w, msg, 404)
			return
		}
		for _, chirp := range chirps {
			respBody := returnValidChirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			}
			returnValidChirps = append(returnValidChirps, respBody)
		}
		dat, errMarshal := json.Marshal(returnValidChirps)
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
	if pathValue != "" && author_id == "" {
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

	if sortKind == "desc" {
		sort.Slice(chirpJSON, func (i, j int) bool { return chirpJSON[i].CreatedAt.After(chirpJSON[j].CreatedAt) })
	} else if sortKind == "asc" {
		log.Println("stick to old sort")
	} else {
		w.WriteHeader(403)
		w.Write([]byte("sort value should be `desc` or `asc`"))
		return
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
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("no bearer")
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("jwt error: %v", err)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
	var postData postDataShape
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&postData)
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
		UserID:    userID,
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
	var postData UserLoginDetail
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&postData)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("JSON decode error: %v", err)))
		return
	}
	hashedPassword, err := auth.HashPassword(postData.Password)
	if err != nil {
		log.Printf("%v\n", err)
		w.WriteHeader(500)
		w.Write([]byte("Server Error"))
		return

	}
	params := database.CreateUserParams{
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          postData.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Printf("failed to create user! %s\n", msg)
		http.Error(w, msg, 500)
		return
	}

	responseJson := returnUser{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
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

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	var postData UserLoginDetail
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&postData)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("JSON decode error: %v", err)))
		return
	}
	var expiresInSeconds time.Duration
	if postData.Expiry == 0 {
		expiresInSeconds = time.Duration(60*60) * time.Second
	} else {
		expiresInSeconds = time.Duration(postData.Expiry) * time.Second
	}
	user, err := cfg.db.GetUser(r.Context(), postData.Email)
	if err != nil {
		log.Printf("%v\n", err)
		http.Error(w, "Unauthorized", 401)
		return
	}
	err = auth.CheckPasswordHash(postData.Password, user.HashedPassword)
	if err != nil {
		log.Printf("%v\n", err)
		http.Error(w, "Unauthorized", 401)
		return
	}

	newJWTToken, err := auth.MakeJWT(user.ID, cfg.tokenSecret, expiresInSeconds)

	if err != nil {
		log.Printf("%v\n", err)
		http.Error(w, "Server Error", 500)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Println(msg)
		http.Error(w, msg, 500)
		return
	}

	responseJson := returnUser{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        newJWTToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	}

	dat, err := json.Marshal(responseJson)
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Println(msg)
		http.Error(w, msg, 500)
		return
	}
	refreshTokenParams := database.AddRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(24*60) * time.Hour),
		UserID:    user.ID,
	}
	_, err = cfg.db.AddRefreshToken(r.Context(), refreshTokenParams)

	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Println(msg)
		http.Error(w, msg, 500)
		return
	}

	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Println("no bearer")
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("jwt error: %v", err)
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
	type updateVal struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var postData updateVal
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&postData)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("JSON decode error: %v", err)))
		return
	}
	hashedPassword, err := auth.HashPassword(postData.Password)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	params := database.UpdateUserDetailsParams{
		Email:          postData.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}
	updatedUser, err := cfg.db.UpdateUserDetails(r.Context(), params)

	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Printf("failed to update user! %s\n", msg)
		http.Error(w, msg, 500)
		return
	}

	responseJson := returnUser{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	}

	dat, err := json.Marshal(responseJson)
	if err != nil {
		msg := fmt.Sprintf("500 - %s", err)
		log.Println(msg)
		http.Error(w, msg, 500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) refreshTheToken(w http.ResponseWriter, r *http.Request) {
	if token, err := auth.GetBearerToken(r.Header); err == nil {
		user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)

		if err != nil {
			msg := fmt.Sprintf("500 - %s", err)
			log.Println(msg)
			http.Error(w, msg, 500)
			return
		}
		expiryParams := database.GetExpiryParams{
			Token:  token,
			UserID: user.ID,
		}

		row, err := cfg.db.GetExpiry(r.Context(), expiryParams)

		if err != nil {
			msg := fmt.Sprintf("500 - %s", err)
			log.Println(msg)
			http.Error(w, msg, 500)
			return
		}

		if row.RevokedAt.Valid {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized"))
			return
		}

		if row.ExpiresAt.Before(time.Now()) {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized"))
			return
		}

		newJWTToken, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Duration(60*60)*time.Second)
		type returnAccessToken struct {
			Token string `json:"token"`
		}

		responseJson := returnAccessToken{
			Token: newJWTToken,
		}

		dat, err := json.Marshal(responseJson)

		if err != nil {
			msg := fmt.Sprintf("500 - %s", err)
			log.Println(msg)
			http.Error(w, msg, 500)
			return
		}
		w.WriteHeader(200)
		w.Write(dat)
		return
	} else {
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	if token, err := auth.GetBearerToken(r.Header); err == nil {
		params := database.RevokeTokenParams{
			Token: token,
			RevokedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
		}
		err = cfg.db.RevokeToken(r.Context(), params)

		if err != nil {
			msg := fmt.Sprintf("500 - %s", err)
			log.Println(msg)
			http.Error(w, msg, 500)
			return
		}
		w.WriteHeader(204)
		return
	} else {
		w.WriteHeader(401)
		w.Write([]byte("Unauthorized"))
		return
	}

}

func (cfg *apiConfig) webhooks(w http.ResponseWriter, r *http.Request) {
	polkaSecret, err := auth.GetApiKey(r.Header)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	if polkaSecret != cfg.polkaSecret {
		w.WriteHeader(401)
		w.Write([]byte("unauthorized"))
		return

	}
	type hook struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	var postData hook
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&postData)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("JSON decode error: %v", err)))
		return
	}
	if postData.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	userID, err := uuid.Parse(postData.Data.UserID)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	user, err := cfg.db.UpgradeUserToChirpyRed(r.Context(), userID)
	if err == nil {
		if user.IsChirpyRed {
			log.Printf("User with email %s was upgraded to chirpy red", user.Email)
			w.WriteHeader(204)
			return
		}
	}
	w.WriteHeader(404)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	tokenSecret := os.Getenv("SIGNING_KEY")
	polkaSecret := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v\n", dbURL, err)
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		db:          dbQueries,
		tokenSecret: tokenSecret,
		polkaSecret: polkaSecret,
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
	mux.Handle("DELETE /api/chirps/{chirpID}", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.chirps)))
	mux.Handle("GET /api/healthz", apiCfg.middlewareMetricsInc(http.HandlerFunc(readiness)))
	mux.Handle("POST /api/users", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.createUser)))
	mux.Handle("PUT /api/users", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.updateUser)))
	mux.Handle("POST /api/login", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.loginUser)))
	mux.Handle("POST /api/revoke", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.revokeToken)))
	mux.Handle("POST /api/refresh", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.refreshTheToken)))
	mux.Handle("POST /api/polka/webhooks", apiCfg.middlewareMetricsInc(http.HandlerFunc(apiCfg.webhooks)))
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
