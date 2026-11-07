package grpc_server

import (
	"context"
	"errors"
	client "task3-http-grpc/client"
	moviepb "task3-http-grpc/proto"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type MovieServiceServer struct {
	moviepb.UnimplementedMovieServiceServer
	Client *client.ExternalClient
}

func (s *MovieServiceServer) LoadMovies(ctx context.Context, req *moviepb.LoadMoviesRequest) (*moviepb.LoadMoviesResponse, error) {

	from := int(req.GetFrom())
	to := int(req.GetTo())

	if from <= 0 || to <= 0 || from > to {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid range: 'from' and 'to' must be positive integers, and 'from' must be less than or equal to 'to'.")
	}

	movies, errorsMap, err := s.Client.FetchMovieFromTo(ctx, from, to)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, status.Errorf(codes.Canceled, "Request was canceled")
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Errorf(codes.DeadlineExceeded, "Request to external service timed out")
		}

		return nil, status.Errorf(codes.Internal, "Failed to fetch movies: %v", err)
	}

	var errorsMapPB map[int32]string
	if len(errorsMap) > 0 {
		errorsMapPB = make(map[int32]string)
		for k, v := range errorsMap {
			errorsMapPB[int32(k)] = v
		}
	}

	var movieResponses []*moviepb.Movie
	for _, movie := range movies {
		movieResponses = append(movieResponses, &moviepb.Movie{
			Id:       int32(movie.MovieId),
			Title:    movie.Title,
			Year:     int32(movie.Year),
			Director: movie.Director,
		})
	}

	return &moviepb.LoadMoviesResponse{
		Movies: movieResponses,
		Errors: errorsMapPB,
	}, nil
}

func (s *MovieServiceServer) GetLastMovie(ctx context.Context, req *moviepb.GetLastMovieRequest) (*moviepb.Movie, error) {
	lastMovie, err := s.Client.FetchLastMovie(ctx)
	if err != nil {

		if errors.Is(err, context.Canceled) {
			return nil, status.Errorf(codes.Canceled, "Request was canceled")
		}

		if errors.Is(err, client.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "No movies found")
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Errorf(codes.DeadlineExceeded, "Request to external service timed out")
		}

		return nil, status.Errorf(codes.Unavailable, "Error fetching last movie: %v", err)

	}

	return &moviepb.Movie{
		Id:       int32(lastMovie.MovieId),
		Title:    lastMovie.Title,
		Year:     int32(lastMovie.Year),
		Director: lastMovie.Director,
	}, nil
}
