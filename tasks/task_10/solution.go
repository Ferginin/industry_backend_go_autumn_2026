package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		clock: clock,
		seq:   0,
		tasks: make(map[string]Task),
	}
}

func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	trimedTitle := strings.TrimSpace(title)
	if trimedTitle == "" {
		return Task{}, ErrInvalidTitle
	}
	r.seq++
	newID := strconv.Itoa(int(r.seq))
	newTask := Task{
		ID:        newID,
		Title:     trimedTitle,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.tasks[newID] = newTask
	return newTask, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tasks[id]
	if !ok {
		return Task{}, false
	}
	return t, true
}

func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resultTasks := []Task{}
	for _, v := range r.tasks {
		if v.Done == done {
			resultTasks = append(resultTasks, v)
		}
	}
	slices.SortFunc(resultTasks, func(a, b Task) int {
		if a.UpdatedAt.After(b.UpdatedAt) {
			return -1
		} else if b.UpdatedAt.After(a.UpdatedAt) {
			return 1
		}
		// можно strings.Compare(a.ID, b.ID)
		if a.ID < b.ID {
			return -1
		} else {
			return 1
		}
	})
	return resultTasks
}

func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if t.Done == done {
		return t, nil
	}
	t.Done = done
	t.UpdatedAt = r.clock.Now()
	r.tasks[id] = t
	return t, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{
		repo: repo,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 2 || len(path) > 3 || path[1] != "tasks" || (len(path) == 3 && path[2] == "") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.URL.Path == "/tasks" {
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
		case http.MethodGet:
			h.handleList(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r, path[2])
	case http.MethodPatch:
		h.handlePatch(w, r, path[2])
	default:
		w.Header().Set("Allow", "GET, PATCH")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	err := decodeStrictJSON(r.Body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task, err := h.repo.Create(req.Title)
	if errors.Is(err, ErrInvalidTitle) {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()["done"]
	done := false
	if len(vals) == 0 {
		done = false
	} else if len(vals) == 1 {
		switch vals[0] {
		case "true":
			done = true
		case "false":
			done = false
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tasks := h.repo.List(done)
	writeJSON(w, http.StatusOK, tasks)
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}
	err := decodeStrictJSON(r.Body, &req)
	if err != nil || req.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task, err := h.repo.SetDone(id, *req.Done)
	if errors.Is(err, ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return err
	}
	err := decoder.Decode(&struct{}{})
	if err == nil {
		return errors.New("unexpected data after json")
	}
	if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// json.NewEncoder(w).Encode(v)
	w.Write(data)
}
