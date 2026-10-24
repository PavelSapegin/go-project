package main

import (
	"log"
	"net/http"
	"os"
	"time"
)




func main(){
	
	standURL := os.Getenv("MOVIE_BASE_URL")
	if standURL == "" {
		standURL = "https://homeworksite.site/"
	}

	serverAddr := os.Getenv("HTTP_ADDR")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	client := &ExcternalClient{}
	client.newClient(standURL, 5*time.Second)

	handler := NewMovieHandler(client)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", handler.GetMovieFromTo)
	mux.HandleFunc("GET /last-movie", handler.GetLastMovie)

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