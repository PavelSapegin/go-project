package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	
	"task3-http-grpc/client"
)

type MovieProvider interface {
	FetchMovie(ctx context.Context, movieId int) (*client.Response, error)
	FetchMovieFromTo(ctx context.Context, from int, to int) ([]*client.Response, error)
	FetchLastMovie(ctx context.Context) (*client.Response, error)
}

type MovieHandler struct {
	Provider MovieProvider
}

func NewMovieHandler(provider MovieProvider) *MovieHandler {
	return &MovieHandler{
		Provider: provider,
	}
}

func (h *MovieHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *MovieHandler) GetMovieFromTo(w http.ResponseWriter, r *http.Request) {
	
	fromStr := strings.TrimSpace(r.URL.Query().Get("from"))
	toStr := strings.TrimSpace(r.URL.Query().Get("to"))
	
	if fromStr == "" || toStr == "" {
		http.Error(w, "Missing 'from' or 'to' parameter", http.StatusBadRequest)
		return
	}

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
		return
	}

	movies, err := h.Provider.FetchMovieFromTo(r.Context(), from, to)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching movies: %v", err), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, http.StatusOK, movies)
} 

func (h *MovieHandler) GetLastMovie(w http.ResponseWriter, r *http.Request) {
	lastMovie, err := h.Provider.FetchLastMovie(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching last movie: %v", err), http.StatusInternalServerError)

		if errors.Is(err, context.Canceled) {
			return
		}

		h.respondJSON(w, http.StatusBadGateway, "External service don't have accessed")
		return 
	}
	h.respondJSON(w, http.StatusOK, lastMovie)
}
