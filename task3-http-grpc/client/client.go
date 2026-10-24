package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrNotFound = errors.New("movie not found")

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

func (c *ExternalClient) FetchMovie(ctx context.Context, movieId int) (*Response, error) {
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
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request failed with status code: %d", resp.StatusCode)
	}

	var movie Response
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, fmt.Errorf("Couldn't decode the response: %w", err)
	}

	return &movie, nil
}

func (c *ExternalClient) FetchMovieFromTo(ctx context.Context, from int, to int) ([]*Response, map[int]string, error) {
	var movies []*Response
	errorsMap := make(map[int]string)

	for i := from; i <= to; i++ {
		movie, err := c.FetchMovie(ctx, i)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, nil, err
			} else {
				errorsMap[i] = err.Error()
				continue
			}
		}
		movies = append(movies, movie)
	}
	return movies, errorsMap, nil
}

func (c *ExternalClient) FetchLastMovie(ctx context.Context) (*Response, error) {
	firstMovie, err := c.FetchMovie(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get first movie: %w", err)
	}

	low := 1
	high := 2

	var lastMovie *Response = firstMovie

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Request canceled: %w", ctx.Err())
		default:
		}

		movie, err := c.FetchMovie(ctx, high)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				break
			}
			return nil, fmt.Errorf("Couldn't get movie with ID %d: %w", high, err)

		}

		lastMovie = movie
		low = high
		high *= 2

	}

	left := low + 1
	right := high - 1

	for left <= right {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Request canceled: %w", ctx.Err())
		default:
		}

		mid := left + (right-left)/2
		movie, err := c.FetchMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				right = mid - 1
				continue
			}
			return nil, fmt.Errorf("Couldn't get movie with ID %d: %w", mid, err)
		}
		lastMovie = movie
		left = mid + 1
	}

	return lastMovie, nil
}
