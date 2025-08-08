package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Uttam1916/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	// get env variables form .env
	godotenv.Load()
	// establish connection to database
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("couldnt establish database connection")
	}
	dbQueries := database.New(db)
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv(dbURL), // from .env
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlderCreateChirps)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlderDisplayAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", apiCfg.handlderDisplaychirp)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func CleanBody(line string) string {

	var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

	for _, word := range profaneWords {
		var wos []string
		line = strings.ToLower(line)
		wos = strings.Split(line, word)
		line = strings.Join(wos, "****")
	}
	return line
}
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type requestBody struct {
		Email string `json:"email"`
	}
	type responseBody struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	var req requestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	resp := responseBody{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) handlderCreateChirps(w http.ResponseWriter, r *http.Request) {
	type reqJson struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	type errorResp struct {
		Error string `json:"error"`
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorResp{Error: "Method not allowed"})
		return
	}

	var req reqJson
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp{Error: "Invalid JSON"})
		return
	}

	req.Body = CleanBody(req.Body)

	if len(req.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp{Error: "Chirp is too long"})
		return
	}

	type respJson struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    string    ` json:"user_id"`
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp{Error: "Invalid user_id"})
		return
	}

	resp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   req.Body,
		UserID: userUUID,
	})

	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(errorResp{Error: fmt.Sprint("%v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(resp)
}

func (cfg *apiConfig) handlderDisplayAllChirps(w http.ResponseWriter, r *http.Request) {
	type errorResp struct {
		Error string `json:"error"`
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorResp{Error: "Method not allowed"})
		return
	}

	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(errorResp{Error: fmt.Sprint("%v", err)})
		return
	}
	w.Header().Set("content-type", "applicaton/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(chirps)
}

func (cfg *apiConfig) handlderDisplaychirp(w http.ResponseWriter, r *http.Request) {
	type errorResp struct {
		Error string `json:"error"`
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorResp{Error: "Method not allowed"})
		return
	}
	chirpID, err := uuid.Parse(r.PathValue("chirp_id"))
	if err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(errorResp{Error: "Invalid chirp id"})
		return
	}
	chirp, err := cfg.db.GetchirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(errorResp{Error: fmt.Sprint("%v", err)})
		return
	}

	w.Header().Set("content-type", "applicaton/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(chirp)
}
