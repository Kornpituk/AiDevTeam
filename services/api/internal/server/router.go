package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/config"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/handler"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/repository"
	_ "github.com/lib/pq"
)

type Server struct {
	cfg    *config.Config
	router *mux.Router
	db     *sql.DB
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func NewServer(cfg *config.Config) *Server {
	connStr := "postgres://" + cfg.Database.User + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + cfg.Database.Port + "/" + cfg.Database.Name + "?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	taskRepo := repository.NewTaskRepository(db)
	eventRepo := repository.NewEventRepository(db)
	artifactRepo := repository.NewArtifactRepository(db)

	taskHandler := handler.NewTaskHandler(taskRepo)
	eventHandler := handler.NewEventHandler(eventRepo)
	artifactHandler := handler.NewArtifactHandler(artifactRepo)

	router := mux.NewRouter()

	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)
	router.Use(corsMiddleware)

	// Health
	router.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Tasks
	router.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	router.HandleFunc("/tasks", taskHandler.GetTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", taskHandler.GetTask).Methods("GET")
	router.HandleFunc("/tasks/{id}/status", taskHandler.UpdateTaskStatus).Methods("PATCH")
	router.HandleFunc("/tasks/{id}/plan", taskHandler.UpdateTaskPlan).Methods("PATCH")
	router.HandleFunc("/tasks/{id}/review-notes", taskHandler.UpdateTaskReviewNotes).Methods("PATCH")

	// Events
	router.HandleFunc("/tasks/{id}/events", eventHandler.CreateEvent).Methods("POST")
	router.HandleFunc("/tasks/{id}/events", eventHandler.GetEvents).Methods("GET")

	// Artifacts
	router.HandleFunc("/tasks/{id}/artifacts", artifactHandler.CreateArtifact).Methods("POST")
	router.HandleFunc("/tasks/{id}/artifacts", artifactHandler.GetArtifacts).Methods("GET")

	return &Server{
		cfg:    cfg,
		router: router,
		db:     db,
	}
}

func (s *Server) Router() *mux.Router {
	return s.router
}
