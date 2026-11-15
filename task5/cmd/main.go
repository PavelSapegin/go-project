package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"context"

	_ "task3-http-grpc/docs"
	moviepb "task3-http-grpc/proto"

	"task3-http-grpc/internal/grpc_server"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"

	"task3-http-grpc/internal/adapter"
	"task3-http-grpc/internal/handler"
	"task3-http-grpc/internal/usecase"
	"task3-http-grpc/internal/repository"
	
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	stdlib "github.com/jackc/pgx/v5/stdlib"
    goose "github.com/pressly/goose/v3"
)

// @title Movie API
// @version 1.0
// @description This is a sample server for a movie API.
// @termsOfService http://swagger.io/terms/
func main() {

	standURL := os.Getenv("MOVIE_BASE_URL")
	if standURL == "" {
		standURL = "https://homeworksite.site"
	}

	serverAddr := os.Getenv("HTTP_ADDR")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	pool, err_pgx := pgxpool.New(ctx, dbURL)
	if err_pgx != nil {
		log.Fatalf("failed to create database pool: %v", err_pgx)
	}
	defer pool.Close()

	err_gpx := pool.Ping(ctx)
	if err_gpx != nil {
		log.Fatalf("failed to ping database: %v", err_gpx)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to open DB from pool %v", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("failed to apply migrations %v", err)
	}

	exClient := adapter.NewExternalClient(standURL, 15*time.Second)
	useCase := usecase.MovieUseCase{MovieProvider: exClient, MovieRepo: repository.NewPostgresMovieRepository(pool)}

	syncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	log.Println("Starting movie synchronization...")
	if err := useCase.SynchronizeMovies(syncCtx); err != nil {
		log.Fatalf("failed to synchronize movies: %v", err)
	}
	log.Println("Movie synchronization completed.")


	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		moviepb.RegisterMovieServiceServer(grpcServer, &grpc_server.MovieServiceServer{UseCase: &useCase})

		log.Printf("Starting gRPC server on %s", ":50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	handler := handler.NewMovieHandler(&useCase)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", handler.GetMovieFromTo)
	mux.HandleFunc("GET /last-movie", handler.GetLastMovie)
	mux.HandleFunc("GET /swagger/", httpSwagger.Handler())

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting server on %s", serverAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", serverAddr, err)
	}
}
