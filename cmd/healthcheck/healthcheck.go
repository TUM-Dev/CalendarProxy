package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	_, err := http.Get("http://127.0.0.1:4321/health")
	if err != nil {
		log.Printf("Healthcheck failed: %s\n", err)
		os.Exit(1)
	}
}
