package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func fetchRequest(ctx context.Context, movieID int, timeout int) (string, error) {

	fullURL := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", movieID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("Can't create a request %w", err)
	}

	client := http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Can't get a response %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Server returned code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Can't read a response %w", err)
	}

	return string(body), nil

}

func main() {
	timeout := 10
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()

	result, err := fetchRequest(ctx, 1, 10)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Println("Result: ", result)
}
