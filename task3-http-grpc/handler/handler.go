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

type LoadMoviesResponse struct {
	Movies []*client.Response `json:"movies"`
	Errors map[int]string     `json:"errors"`
}

type ErrorResponse struct {
	Error string `json:"message"`
}

type MovieProvider interface {
	FetchMovie(ctx context.Context, movieId int) (*client.Response, error)
	FetchMovieFromTo(ctx context.Context, from int, to int) ([]*client.Response, map[int]string, error)
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

// GetMovieFromTo handles the GET /load-movies endpoint to fetch movies from a specified range.
// @Summary Get movies from a specified range
// @Description Fetches movies from the external service based on the provided 'from' and 'to' query parameters.
// @Produce json
// @Param from query int true "Starting movie ID"
// @Param to query int true "Ending movie ID"
// @Success 200 {object} LoadMoviesResponse "List of fetched movies along with any partial errors"
// @Failure 400 {string} string "Missing or invalid 'from' or 'to' parameter"
// @Failure 500 {string} string "Internal server error"
// @Failure 504 {string} string "Request to external service timed out"
// @Router /load-movies [get]
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

	if from <= 0 || to <= 0 || from > to {
		http.Error(w, "Invalid range: 'from' and 'to' must be positive integers, and 'from' must be less than or equal to 'to'.", http.StatusBadRequest)
		return
	}

	movies, errorsMap, err := h.Provider.FetchMovieFromTo(r.Context(), from, to)
	if err != nil {

		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Request to external service timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, fmt.Sprintf("Error fetching movies: %v", err), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]any{
		"movies": movies,
		"errors": errorsMap,
	})
} 

// @Summary Get the last movie
// @Description Fetches the last movie from the external service.
// @Produce json
// @Success 200 {object} client.Response "Last movie details"
// @Failure 404 {string} string "No movies found"
// @Failure 502 {object} ErrorResponse "External service error"
// @Failure 504 {string} string "Request to external service timed out"
// @Router /last-movie [get]
func (h *MovieHandler) GetLastMovie(w http.ResponseWriter, r *http.Request) {
	lastMovie, err := h.Provider.FetchLastMovie(r.Context())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, client.ErrNotFound) {
			http.Error(w, "No movies found", http.StatusNotFound)
			return
		}



		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Request to external service timed out", http.StatusGatewayTimeout)
			return
		}
		h.respondJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("Error fetching last movie: %v", err)})
		return 
	}
	h.respondJSON(w, http.StatusOK, lastMovie)
}
