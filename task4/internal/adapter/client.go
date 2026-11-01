package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"task3-http-grpc/internal/domain"
)


type Response struct {
	MovieId  int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type ExternalClient struct {
	BaseURL string
	client  *http.Client
}

func NewExternalClient(url string, timeout time.Duration) *ExternalClient {
	c := &ExternalClient{}
	c.BaseURL = url
	c.client = &http.Client{
		Timeout: timeout,
	}
	return c
}

func (c *ExternalClient) GetMovie(ctx context.Context, movieId int) (*domain.Movie, error) {
	endpoint, err := url.JoinPath(c.BaseURL, fmt.Sprintf("%d", movieId), "info.0.json")
	if err != nil {
		return nil, fmt.Errorf("Couldn't create a URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't create a request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Couldn't send a request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request failed with status code: %d", resp.StatusCode)
	}

	var movie Response
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, fmt.Errorf("Couldn't decode the response: %w", err)
	}

	return &domain.Movie{
		MovieId:  movie.MovieId,
		Title:    movie.Title,
		Year:     movie.Year,
		Director: movie.Director,
	}, nil
}



