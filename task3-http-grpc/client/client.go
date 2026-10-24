package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"errors"
)

var ErrNotFound = errors.New("movie not found")

type Response struct {
	MovieId int`json:"id"`
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

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	
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
	/* 
	считаем, что фильмы на стенде имеют id подряд от 1 до last_id; (если комикс с id=99999 не найден (404), то комиксов с большим id гарантированно не будет)
	ищем по типу бинарного поиска последний фильм, который существует на стенде (не возвращает 404)
	*/

	firstMovie, err := c.FetchMovie(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("Could'nt get first movie: %w", err)
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
			return nil, fmt.Errorf("Could'nt get movie with ID %d: %w", high, err)
		
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
		
		mid := left + (right - left) / 2
		movie, err := c.FetchMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				right = mid - 1
				continue
			}
			return nil, fmt.Errorf("Could'nt get movie with ID %d: %w", mid, err)
		}
		lastMovie = movie
		left = mid + 1
	}

	return lastMovie, nil
}
