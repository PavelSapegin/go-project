package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"

	domain "task3-http-grpc/internal/domain"
)

type PostgresMovieRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMovieRepository(pool *pgxpool.Pool) *PostgresMovieRepository {
	return &PostgresMovieRepository{pool: pool}
}

func (r *PostgresMovieRepository) GetLastMovie(ctx context.Context) (*domain.Movie, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, title, year, director FROM movies ORDER BY id DESC LIMIT 1")
	var movie domain.Movie
	err := row.Scan(&movie.MovieId, &movie.Title, &movie.Year, &movie.Director)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &movie, nil
}

func (r *PostgresMovieRepository) GetMoviesBatch(ctx context.Context, from int, to int) ([]*domain.Movie, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, title, year, director FROM movies WHERE id BETWEEN $1 AND $2", from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*domain.Movie

	for rows.Next() {
		var movie domain.Movie
		err := rows.Scan(&movie.MovieId, &movie.Title, &movie.Year, &movie.Director)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrNotFound
			}
			return nil, err
		}
		movies = append(movies, &movie)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *PostgresMovieRepository) SaveMovies(ctx context.Context, movies []*domain.Movie) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	
	for _, movie := range movies {
		_, err := tx.Exec(ctx, "INSERT INTO movies (id, title, year, director) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING", movie.MovieId, movie.Title, movie.Year, movie.Director)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit(ctx)
}

func (r *PostgresMovieRepository) GetAllMovies(ctx context.Context) ([]*domain.Movie, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, title, year, director FROM movies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	movies := []*domain.Movie{}
	for rows.Next() {
		var movie domain.Movie
		err := rows.Scan(&movie.MovieId, &movie.Title, &movie.Year, &movie.Director)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *PostgresMovieRepository) GetMovieByIDs(ctx context.Context, movieIDs []int) ([]*domain.Movie, error) {
	movies := make([]*domain.Movie, 0, len(movieIDs))
	if len(movieIDs) == 0 {
		return movies, nil
	}

	rows, err := r.pool.Query(ctx, "SELECT id, title, year, director FROM movies WHERE id = ANY($1)", movieIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie domain.Movie
		err := rows.Scan(&movie.MovieId, &movie.Title, &movie.Year, &movie.Director)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}
