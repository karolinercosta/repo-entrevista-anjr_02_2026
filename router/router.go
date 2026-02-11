package router

import (
	"context"
	"os"
	"time"

	"github.com/gorilla/mux"

	"example.com/tasksapi/handlers"
	"example.com/tasksapi/models"
	"example.com/tasksapi/store"
)

func New() *mux.Router {
	logger := models.NewDefaultLogger()
	return NewWithLogger(logger)
}

func NewWithLogger(logger models.Logger) *mux.Router {
	var s store.Store

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "taskdb"
	}

	collectionName := os.Getenv("MONGO_COLLECTION")
	if collectionName == "" {
		collectionName = "tasks"
	}

	// Tenta conexao para o MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	m, err := store.NewMongo(ctx, mongoURI, dbName, collectionName)
	cancel()

	if err != nil {
		logger.Warn("failed to connect to MongoDB (%s): %v", mongoURI, err)
		logger.Info("falling back to in-memory store")
		s = store.New()
	} else {
		s = m
		logger.Info("successfully connected to MongoDB: %s", mongoURI)
	}

	api := handlers.NewAPI(s, logger)
	r := mux.NewRouter()
	r.HandleFunc("/tasks", api.CreateTask).Methods("POST")
	r.HandleFunc("/tasks", api.ListTasks).Methods("GET")
	r.HandleFunc("/tasks/{id}", api.GetTask).Methods("GET")
	r.HandleFunc("/tasks/{id}", api.UpdateTask).Methods("PUT")
	r.HandleFunc("/tasks/{id}", api.DeleteTask).Methods("DELETE")
	return r
}
