package server

import (
	"database/sql"
	"log"

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

func NewServer(cfg *config.Config) *Server {
	db, err := sql.Open("postgres", cfg.Database.User+":"+cfg.Database.Password+"@"+cfg.Database.Host+":"+cfg.Database.Port+"/"+cfg.Database.Name+"?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

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

	// Health
	router.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Tasks
	router.HandleFunc("/tasks", taskHandler.CreateTask).Methods("POST")
	router.HandleFunc("/tasks", taskHandler.GetTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", taskHandler.GetTask).Methods("GET")
	router.HandleFunc("/tasks/{id}/status", taskHandler.UpdateTaskStatus).Methods("PATCH")

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