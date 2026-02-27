package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/neocomx/TaskQueueService/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neocomx/TaskQueueService/worker"
)

func setupServer() (*Server, *task.Store, *worker.Pool) {
	store := task.NewStore()
	processor := &worker.PrintProcessor{}
	pool := worker.NewPool(1, store, processor, 5*time.Second)
	pool.Start()
	server := NewServer(store, pool)
	return server, store, pool
}

func TestHandleTasks_POST(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
	}{
		{
			name:           "creates task correctly",
			body:           map[string]string{"payload": "hello"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "fails if payload is empty",
			body:           map[string]string{"payload": ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "fails if body is invalid",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _, pool := setupServer()
			defer pool.Shutdown()

			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			} else {
				bodyBytes = []byte("invalid json{")
			}

			req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux := http.NewServeMux()
			server.RegisterRoutes(mux)
			mux.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestHandleTasks_GET_List(t *testing.T) {
	server, store, pool := setupServer()
	defer pool.Shutdown()

	store.Save("task 1")
	store.Save("task 2")

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var tasks []*task.Task
	err := json.NewDecoder(rec.Body).Decode(&tasks)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestHandleTaskByID_GET(t *testing.T) {
	server, store, pool := setupServer()
	defer pool.Shutdown()

	saved := store.Save("hello")

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+saved.ID, nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var found task.Task
	err := json.NewDecoder(rec.Body).Decode(&found)
	require.NoError(t, err)
	assert.Equal(t, saved.ID, found.ID)
}

func TestHandleTaskByID_NotFound(t *testing.T) {
	server, _, pool := setupServer()
	defer pool.Shutdown()

	req := httptest.NewRequest(http.MethodGet, "/tasks/id-notfound", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var body map[string]string
	err := json.NewDecoder(rec.Body).Decode(&body)
	require.NoError(t, err)
	assert.Contains(t, body["error"], "not found")
}

func TestHandleTasks_MethodNotAllowed(t *testing.T) {
	server, _, pool := setupServer()
	defer pool.Shutdown()

	req := httptest.NewRequest(http.MethodDelete, "/tasks", nil)
	rec := httptest.NewRecorder()

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
