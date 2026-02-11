package main

import (
	"net/http"

	"example.com/tasksapi/models"
	"example.com/tasksapi/router"
	"github.com/rs/cors"
)

func main() {
	logger := models.NewDefaultLogger()
	r := router.New()
	c := cors.AllowAll() // For development, allow all origins
	logger.Info("starting server on :8080")
	if err := http.ListenAndServe(":8080", c.Handler(r)); err != nil {
		logger.Fatal("server failed: %v", err)
	}
}
