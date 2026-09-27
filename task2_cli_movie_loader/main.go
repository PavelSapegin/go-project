package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"sync"
)

type Response struct {
	MovieID int `json:"id"`
	Title string `json:"title"`
	Year int `json:"year"`
	Director string `json:"director"`
}


func fetchRequest(ctx context.Context, client *http.Client, movieID int, timeout int) (Response, error) {	
	
	fullURL := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", movieID)
	var parsed_resp Response

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return parsed_resp, fmt.Errorf("Can't create a request %w", err)
	}

	resp, err := (*client).Do(req)
	if err != nil {
		return parsed_resp, fmt.Errorf("Can't get a response %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parsed_resp, fmt.Errorf("Server returned code %d", resp.StatusCode)
	}

	
	err = json.NewDecoder(resp.Body).Decode(&parsed_resp)
	if err != nil {
		return parsed_resp, fmt.Errorf("Can't parsed response %w\n", err)
	}

	return parsed_resp, nil
}


func worker(ctx context.Context, wg *sync.WaitGroup, client *http.Client, id int, jobs <-chan int, results chan<- Response, timeout int) {

	defer wg.Done()
	for {
		select {
		case <- ctx.Done():
			return
		
		case movieID, ok := <- jobs:

			if !ok {
				return
			}

			res, err := fetchRequest(ctx, client, movieID, timeout)
			if err != nil {
				fmt.Printf("Failed to fetch request: %v\n", err)
				continue
			}

			select {
			case <- ctx.Done():
				fmt.Printf("Context for worker %d was canceled\n", id)
				return
			
			case results <- res:
			}

		}
	}
}

func main() {
	timeout := 10
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()

	client := http.Client{}

	var wg sync.WaitGroup


	jobs := make(chan int,10)
	results := make(chan Response, 10)

	for id := range(10) {
		wg.Add(1)
		go worker(ctx, &wg, &client, id, jobs, results, timeout)
	}

	for id := 1; id <= 10; id++ {
		jobs <- id
	}

	close(jobs)
	wg.Wait()
	close(results)
	

	for result := range results{
		fmt.Println("Result: ")
		fmt.Println("Id:" ,result.MovieID)
		fmt.Println("Title: ", result.Title)
		fmt.Println("Year: ", result.Year)
		fmt.Println("Director: ", result.Director)
		fmt.Println("-----------------")
	}

}
