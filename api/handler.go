package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/neocomx/TaskQueueService/task"
	"github.com/neocomx/TaskQueueService/worker"
)

type Server struct {
	store *task.Store
	pool  *worker.Pool
}

func NewServer(store *task.Store, pool *worker.Pool) *Server {
	return &Server{store: store, pool: pool}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/tasks", s.handleTasks)
	mux.HandleFunc("/tasks/", s.handleTaskByID)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createTask(w, r)
	case http.MethodGet:
		s.listTasks(w, r)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Payload string `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if body.Payload == "" {
		writeError(w, "payload is required", http.StatusBadRequest)
		return
	}

	t := s.store.Save(body.Payload)
	s.pool.Submit(t)

	writeJSON(w, t, http.StatusCreated)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.store.List()
	writeJSON(w, tasks, http.StatusOK)
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		writeError(w, "task id is required", http.StatusBadRequest)
		return
	}

	t, err := s.store.Get(id)
	if err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, t, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, map[string]string{"error": msg}, status)
}
