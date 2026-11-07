package usecase

import (
	"errors"
	"context"
	"fmt"

	"task3-http-grpc/internal/domain"
)

type MovieUseCase struct {
	MovieProvider domain.MovieProvider
}

func (m *MovieUseCase) LoadMovies(ctx context.Context, from int, to int) ([]*domain.Movie, map[int]string, error) {
	var movies []*domain.Movie
	errorsMap := make(map[int]string)

	if from <= 0 || to <= 0 || from > to {
		return nil, nil, domain.ErrInvalidRange
	}

	for i := from; i <= to; i++ {
		movie, err := m.MovieProvider.GetMovie(ctx, i)
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




func (m *MovieUseCase) GetLastMovie(ctx context.Context) (*domain.Movie, error) {
	firstMovie, err := m.MovieProvider.GetMovie(ctx, 1)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get first movie: %w", err)
	}

	low := 1
	high := 2

	var lastMovie *domain.Movie = firstMovie

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Request canceled: %w", ctx.Err())
		default:
		}

		movie, err := m.MovieProvider.GetMovie(ctx, high)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
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
		movie, err := m.MovieProvider.GetMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
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
