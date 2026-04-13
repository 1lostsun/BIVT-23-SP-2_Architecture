package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"lab2/internal/repo/pg"
)

type UseCase interface {
	GetNotes(ctx context.Context) ([]pg.Note, error)
	CreateNote(ctx context.Context, title, body string) (pg.Note, error)
	DeleteNote(ctx context.Context, id int) error
}

type Handler struct {
	uc UseCase
}

func New(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /notes", h.getNotes)
	mux.HandleFunc("POST /notes", h.createNote)
	mux.HandleFunc("DELETE /notes/{id}", h.deleteNote)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) getNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.uc.GetNotes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

type createRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	note, err := h.uc.CreateNote(r.Context(), req.Title, req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.uc.DeleteNote(r.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
