package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/config"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/handler"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/llm"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/repository"
	"github.com/Kornpituk/AiDevTeam/services/api/internal/service"
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
	profileRepo := repository.NewAgentProfileRepository(db)
	teamRepo := repository.NewAgentTeamRepository(db)
	teamMemberRepo := repository.NewAgentTeamMemberRepository(db)
	runRepo := repository.NewAgentRunRepository(db)
	stepRepo := repository.NewAgentRunStepRepository(db)
	messageRepo := repository.NewAgentMessageRepository(db)
	approvalRepo := repository.NewHumanApprovalRepository(db)
	toolCallRepo := repository.NewAgentToolCallRepository(db)

	llmProvider, err := llm.NewProviderFromConfig(cfg.LLM)
	if err != nil {
		log.Printf("Warning: failed to initialize LLM provider: %v, using fake provider", err)
		llmProvider = llm.NewFakeProvider()
	}

	orchestratorRepo := service.NewOrchestratorRepoImpl(runRepo, stepRepo, teamMemberRepo, profileRepo, messageRepo)
	orchestrator := service.NewOrchestrator(orchestratorRepo, llmProvider)

	taskHandler := handler.NewTaskHandler(taskRepo)
	eventHandler := handler.NewEventHandler(eventRepo)
	artifactHandler := handler.NewArtifactHandler(artifactRepo)
	profileHandler := handler.NewAgentProfileHandler(profileRepo)
	teamHandler := handler.NewAgentTeamHandler(teamRepo, teamMemberRepo)
	runHandler := handler.NewAgentRunHandler(runRepo, orchestrator)
	stepHandler := handler.NewAgentRunStepHandler(stepRepo)
	messageHandler := handler.NewAgentMessageHandler(messageRepo)
	approvalHandler := handler.NewHumanApprovalHandler(approvalRepo)
	toolCallHandler := handler.NewAgentToolCallHandler(toolCallRepo)

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

	// Agent Profiles
	router.HandleFunc("/agent-profiles", profileHandler.CreateAgentProfile).Methods("POST")
	router.HandleFunc("/agent-profiles", profileHandler.GetAgentProfiles).Methods("GET")
	router.HandleFunc("/agent-profiles/{id}", profileHandler.GetAgentProfile).Methods("GET")

	// Agent Teams
	router.HandleFunc("/agent-teams", teamHandler.CreateAgentTeam).Methods("POST")
	router.HandleFunc("/agent-teams", teamHandler.GetAgentTeams).Methods("GET")
	router.HandleFunc("/agent-teams/{id}", teamHandler.GetAgentTeam).Methods("GET")
	router.HandleFunc("/agent-teams/{id}/members", teamHandler.CreateTeamMember).Methods("POST")
	router.HandleFunc("/agent-teams/{id}/members", teamHandler.GetTeamMembers).Methods("GET")

	// Agent Runs
	router.HandleFunc("/tasks/{id}/agent-runs", runHandler.CreateAgentRun).Methods("POST")
	router.HandleFunc("/tasks/{id}/agent-runs", runHandler.GetAgentRunsByTask).Methods("GET")
	router.HandleFunc("/agent-runs/{id}", runHandler.GetAgentRun).Methods("GET")
	router.HandleFunc("/agent-runs/{id}/start", runHandler.StartAgentRun).Methods("POST")
	router.HandleFunc("/agent-runs/{id}/cancel", runHandler.CancelAgentRun).Methods("POST")

	// Agent Run Steps
	router.HandleFunc("/agent-runs/{id}/steps", stepHandler.CreateRunStep).Methods("POST")
	router.HandleFunc("/agent-runs/{id}/steps", stepHandler.GetRunSteps).Methods("GET")
	router.HandleFunc("/agent-run-steps/{id}/status", stepHandler.UpdateStepStatus).Methods("PATCH")

	// Agent Messages
	router.HandleFunc("/agent-runs/{id}/messages", messageHandler.CreateMessage).Methods("POST")
	router.HandleFunc("/agent-runs/{id}/messages", messageHandler.GetMessages).Methods("GET")

	// Human Approvals
	router.HandleFunc("/agent-runs/{id}/approvals", approvalHandler.CreateApproval).Methods("POST")
	router.HandleFunc("/agent-runs/{id}/approvals", approvalHandler.GetApprovals).Methods("GET")
	router.HandleFunc("/human-approvals/{id}/status", approvalHandler.UpdateApprovalStatus).Methods("PATCH")

	// Agent Tool Calls
	router.HandleFunc("/agent-runs/{id}/tool-calls", toolCallHandler.CreateToolCall).Methods("POST")
	router.HandleFunc("/agent-runs/{id}/tool-calls", toolCallHandler.GetToolCalls).Methods("GET")
	router.HandleFunc("/agent-tool-calls/{id}/status", toolCallHandler.UpdateToolCallStatus).Methods("PATCH")

	return &Server{
		cfg:    cfg,
		router: router,
		db:     db,
	}
}

func (s *Server) Router() *mux.Router {
	return s.router
}
