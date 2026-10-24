package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "task3-http-grpc/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"

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