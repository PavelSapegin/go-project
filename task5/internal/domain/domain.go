package domain

import (
	"errors"
	"context"
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
}

var ErrNotFound = errors.New("movie not found")
var ErrInvalidRange = errors.New("invalid range: 'from' and 'to' must be positive integers, and 'from' must be less than or equal to 'to'")