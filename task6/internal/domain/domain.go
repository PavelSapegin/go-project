package domain

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
)

type Movie struct {
	MovieId  int
	Title    string
	Year     int
	Director string
}

type MovieProvider interface {
	GetMovie(ctx context.Context, movieId int) (*Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from int, to int) ([]*Movie, map[int]string, error)
	GetLastMovie(ctx context.Context) (*Movie, error)
	SearchMovies(ctx context.Context, query string) ([]*Movie, error)
}

type MovieRepository interface {
	GetLastMovie(ctx context.Context) (*Movie, error)
	GetMoviesBatch(ctx context.Context, from int, to int) ([]*Movie, error)
	SaveMovies(ctx context.Context, movies []*Movie) error
	GetAllMovies(ctx context.Context) ([]*Movie, error)
	GetMovieByIDs(ctx context.Context, movieIDs []int) ([]*Movie, error)
}

type SearchIndex interface {
	IndexMovies(ctx context.Context, movies []*Movie) error
	SearchIDsByWords(ctx context.Context, keywords []string) ([]int, error)
}

type SearchCache interface {
	GetSearchResult(ctx context.Context, key string) ([]*Movie, error)
	SaveSearchResult(ctx context.Context, key string, movies []*Movie, ttl time.Duration) error
	CleanCache(ctx context.Context) error
}
var ErrNotFound = errors.New("movie not found")
var ErrInvalidRange = errors.New("invalid range: 'from' and 'to' must be positive integers, and 'from' must be less than or equal to 'to'")
var ErrEmptyQuery = errors.New("search query cannot be empty")
var ErrCacheMiss = errors.New("cache miss: no movies found for the given word")


func Tokenize(text string) []string {
	textLower := strings.ToLower(text)
	words := strings.FieldsFunc(textLower, func(r rune) bool {return !unicode.IsLetter(r) && !unicode.IsDigit(r)})
	return words
}