package grpc_server

import (
	"context"
	moviepb "task3-http-grpc/proto"
	client "task3-http-grpc/client"
)

type MovieServiceServer struct {
	moviepb.UnimplementedMovieServiceServer
	Client *client.ExcternalClient
}

func (s *MovieServiceServer) LoadMovies(ctx context.Context, req *moviepb.LoadMoviesRequest) (*moviepb.LoadMoviesResponse, error) {

	from := int(req.GetFrom())
	to := int(req.GetTo())

	movies, err := s.Client.FetchMovieFromTo(ctx, from, to)
	if err != nil {
		return nil, err
	}
	
	var movieResponses []*moviepb.Movie
	for _, movie := range movies {
		movieResponses = append(movieResponses, &moviepb.Movie{
			Id: int32(movie.MovieId),
			Title:   movie.Title,
			Year:    int32(movie.Year),
			Director: movie.Director,
		})
	}

	return &moviepb.LoadMoviesResponse{
		Movies: movieResponses,
	}, nil
}

func (s *MovieServiceServer) LoadLastMovie(ctx context.Context, req *moviepb.LoadLastMovieRequest) (*moviepb.Movie, error) {
	lastMovie, err := s.Client.FetchLastMovie(ctx)
	if err != nil {
		return nil, err
	}

	return &moviepb.Movie{
		Id: int32(lastMovie.MovieId),
		Title:   lastMovie.Title,
		Year:    int32(lastMovie.Year),
		Director: lastMovie.Director,
	}, nil
}