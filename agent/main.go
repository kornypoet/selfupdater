package main

import (
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/mod/semver"
)

//go:embed VERSION
var version string

type ApiResponse struct {
	Latest string `json:"latest"`
}

func main() {
	for {
		log.Printf("Running version: %s\n", version)
		time.Sleep(1 * time.Second) // perform "work"
		log.Printf("Update available? %v", updateAvailable())
	}
}

func updateAvailable() bool {
	url := "http://localhost:8080/latest"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var data ApiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	log.Printf("Response:\n%+v\n", data)
	if semver.Compare(version, data.Latest) == -1 {
		return true
	}
	return false
}
