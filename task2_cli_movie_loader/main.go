package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"sync"
	"flag"
	"slices"
	"cmp"
	"os"
)

type Response struct {
	MovieID int `json:"id"`
	Title string `json:"title"`
	Year int `json:"year"`
	Director string `json:"director"`
}

func exitWithError(msg string) {
	fmt.Printf("Error in arguments: %s\n", msg)
	flag.Usage()
	os.Exit(1)
}

func fetchRequest(ctx context.Context, client *http.Client, movieID int) (Response, error) {	
	
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


func worker(ctx context.Context, wg *sync.WaitGroup, client *http.Client, id int, jobs <-chan int, results chan<- Response) {

	defer wg.Done()
	for {
		select {
		case <- ctx.Done():
			return
		
		case movieID, ok := <- jobs:

			if !ok {
				return
			}

			res, err := fetchRequest(ctx, client, movieID)
			if err != nil {
				fmt.Printf("Failed to fetch request for movieID = %d: %v\n", movieID, err)
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

	fromPtr := flag.Int("from",-1, "from (required) >= 0")
	toPtr := flag.Int("to",-1,"to (required) >= 0")
	workersNumPtr := flag.Int("workers", 10, "nums of workers >= 0")
	timeoutPtr := flag.Int("timeout", 5, "timeout for HTTP-request >= 0")

	flag.Parse()

	from := *fromPtr
	to := *toPtr
	workersNum := *workersNumPtr
	timeout := *timeoutPtr

	if from == -1 || to == -1 {
		exitWithError("Params from and to required")
	}

	if from <= 0 {
		exitWithError("Param from must be greater than 1")
	}

	if to <= 0 {
		exitWithError("Param to must be greater than 1")
	}

	if from >= to {
		exitWithError("Param to must be greater than from")
	}

	if workersNum <= 0 {
		exitWithError("Param workers must be greater than 1")
	}

	if timeout <= 0 {
		exitWithError("Param timeout must be greater than 0")
	}

	

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()

	client := http.Client{}

	var wg sync.WaitGroup

	num_jobs := to - from + 1
	jobs := make(chan int, num_jobs)
	results := make(chan Response, num_jobs)

	for id := range(workersNum) {
		wg.Add(1)
		go worker(ctx, &wg, &client, id, jobs, results)
	}

	for id := from; id <= to; id++ {
		jobs <- id
	}

	close(jobs)
	wg.Wait()
	close(results)


	results_arr := make([]Response, 0, num_jobs)
	for results := range results{
		results_arr = append(results_arr, results)
	}

	slices.SortFunc(results_arr, func(a, b Response) int {
		return cmp.Compare(a.MovieID, b.MovieID)
	})

	for _, result := range results_arr{
		fmt.Println("Result: ")
		fmt.Println("Id:" ,result.MovieID)
		fmt.Println("Title: ", result.Title)
		fmt.Println("Year: ", result.Year)
		fmt.Println("Director: ", result.Director)
		fmt.Println("-----------------")
	}

}
