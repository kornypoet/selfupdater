package main

import (
	_ "embed"
	"log"
	"time"
)

//go:embed VERSION
var version string

func main() {
	for {
		log.Printf("Running version: %s\n", version)
		time.Sleep(5 * time.Second)
	}
}
