package main

import (
    "log"
    "net/http"

    "example.com/tasksapi/router"
)

func main() {
    r := router.New()
    log.Println("starting server on :8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
