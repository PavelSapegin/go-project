package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"sync"

	"task3-http-grpc/internal/domain"
)

type MovieUseCase struct {
	MovieProvider domain.MovieProvider
	MovieRepo domain.MovieRepository
}

func (m *MovieUseCase) LoadMovies(ctx context.Context, from int, to int) ([]*domain.Movie, map[int]string, error) {

	if from <= 0 || to <= 0 || from > to {
		return nil, nil, domain.ErrInvalidRange
	}

	movies, err := m.MovieRepo.GetMoviesBatch(ctx, from, to)
	if err != nil {
		return nil, nil, fmt.Errorf("Couldn't get movies batch from repository: %w", err)
	}

	return movies, nil, nil
}

func (m *MovieUseCase) findLastExternalMovie(ctx context.Context) (*domain.Movie, error) {
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

func (m *MovieUseCase) GetLastMovie(ctx context.Context) (*domain.Movie, error) {
	lastMovie, err := m.MovieRepo.GetLastMovie(ctx)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get last movie from repository: %w", err)
	}

	return lastMovie, nil
}

func worker(ctx context.Context, wg *sync.WaitGroup,jobs <-chan int, results chan<- *domain.Movie, m *MovieUseCase, timeout int) {
	defer wg.Done()
	for {
		select {
		case <- ctx.Done():
			return

		case movieID, ok := <- jobs:

			if !ok {
				return
			}

			workerCtx, workerCancel := context.WithTimeout(ctx, time.Duration(timeout) * time.Second)
			res, err := m.MovieProvider.GetMovie(workerCtx, movieID)
			workerCancel()

			if err != nil {
				fmt.Printf("Failed to fetch request for movieID = %d: %v\n", movieID, err)
				continue
			}
			
			select {
			case <- ctx.Done():
				fmt.Printf("Context for worker %d was canceled\n", movieID)
				return
			
			case results <- res:
				fmt.Printf("Get movie %d\n", movieID)
			}
		}

	}
}

func (m *MovieUseCase) SynchronizeMovies(ctx context.Context) error {
	lastMovie, err := m.findLastExternalMovie(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // nothing to sync if there are no movies
		}
		return fmt.Errorf("Couldn't get last movie: %w", err)
	}

	fmt.Printf("Found last film")

	existingMovies, err := m.MovieRepo.GetMoviesBatch(ctx, 1, lastMovie.MovieId)
	if err != nil {
		return fmt.Errorf("Couldn't get existing movies from repository: %w", err)
	}

	existingMovieMap := make(map[int]struct{}, len(existingMovies))
	for _, movie := range existingMovies {
		existingMovieMap[movie.MovieId] = struct{}{}
	}

	jobs := make(chan int)
	results := make(chan *domain.Movie)

	wg := sync.WaitGroup{}

	for range 20 {
		wg.Add(1)
		go worker(ctx, &wg, jobs, results, m, 5)
	}
	
	go func() {
		defer close(jobs)
		for id := 1; id <= lastMovie.MovieId; id++ {
			if _, exists := existingMovieMap[id]; !exists {
				select {
				case <-ctx.Done():
					return
				case jobs <- id:
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()
	
	var moviesToSave []*domain.Movie
	for movie := range results {
		moviesToSave = append(moviesToSave, movie)
	}

	if len(moviesToSave) > 0 {
		err = m.MovieRepo.SaveMovies(ctx, moviesToSave)
		if err != nil {
			return fmt.Errorf("Couldn't save movies to repository: %w", err)
		}
	}

	return nil
}
