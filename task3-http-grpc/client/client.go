package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Response struct {
	MovieId int`json:"movie_id"`
	Title string `json:"title"`
	Year int `json:"year"`
	Director string `json:"director"`
}

type ExcternalClient struct {
	BaseURL string
	client *http.Client
}


func (c *ExcternalClient) NewClient(url string, timeout time.Duration) *http.Client {
	c.BaseURL = url
	c.client = &http.Client{
		Timeout: timeout,
	}
	return c.client
}

func (c *ExcternalClient) FetchMovie(ctx context.Context, movieId int) (*Response, error) {
	endpoint, err := url.JoinPath(c.BaseURL, fmt.Sprintf("%d", movieId), "info.0.json")
	if err != nil {
		return nil, fmt.Errorf("Could'nt create a URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Could'nt create a request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could'nt send a request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request failed with status code: %d", resp.StatusCode)
	}

	var movie Response
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, fmt.Errorf("Could'nt decode the response: %w", err)
	}

	return &movie, nil
}

func (c *ExcternalClient) FetchMovieFromTo(ctx context.Context, from int, to int) ([]*Response, error) {
	var movies []*Response
	for i := from; i <= to; i++ {
		movie, err := c.FetchMovie(ctx, i)
		if err != nil {
			return nil, fmt.Errorf("Could'nt get movie with ID %d: %w", i, err)
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

func (c *ExcternalClient) FetchLastMovie(ctx context.Context) (*Response, error) {
	
	var lastMovie *Response
	for i := 1; ; i++ {
		movie, err := c.FetchMovie(ctx, i)
		if err != nil {
			if httpErr, ok := err.(*url.Error); ok && httpErr.Timeout() {
				return nil, fmt.Errorf("Request timed out: %w", err)
			}
			if httpErr, ok := err.(*url.Error); ok && httpErr.Err.Error() == "404 Not Found" {
				break
			}
			return nil, fmt.Errorf("Could'nt get movie with ID %d: %w", i, err)
		}
		lastMovie = movie
	}
	if lastMovie == nil {
		return nil, fmt.Errorf("No movies found")
	}
	return lastMovie, nil
}
