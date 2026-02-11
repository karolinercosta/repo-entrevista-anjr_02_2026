package main

import (
	"log"
	"net/http"

	"example.com/tasksapi/router"
	"github.com/rs/cors"
)

func main() {
	r := router.New()
	c := cors.AllowAll() // For development, allow all origins
	log.Println("starting server on :8080")
	if err := http.ListenAndServe(":8080", c.Handler(r)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
