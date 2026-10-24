package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"net"

	_ "task3-http-grpc/docs"
	moviepb "task3-http-grpc/proto"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"google.golang.org/grpc"
	"task3-http-grpc/grpc_server"

	"task3-http-grpc/client"
	"task3-http-grpc/handler"
)

// @title Movie API
// @version 1.0
// @description This is a sample server for a movie API.
// @termsOfService http://swagger.io/terms/
func main(){
	
	standURL := os.Getenv("MOVIE_BASE_URL")
	if standURL == "" {
		standURL = "https://homeworksite.site/"
	}

	serverAddr := os.Getenv("HTTP_ADDR")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	exClient := client.ExcternalClient{}
	exClient.NewClient(standURL, 5*time.Second)

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		moviepb.RegisterMovieServiceServer(grpcServer, &grpc_server.MovieServiceServer{Client: &exClient})
		
		log.Printf("Starting gRPC server on %s", ":50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	
	handler := handler.NewMovieHandler(&exClient)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", handler.GetMovieFromTo)
	mux.HandleFunc("GET /last-movie", handler.GetLastMovie)
	mux.HandleFunc("GET /swagger/", httpSwagger.Handler())

	server := &http.Server{
		Addr: serverAddr,
		Handler: mux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	log.Printf("Starting server on %s", serverAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", serverAddr, err)
	}
}