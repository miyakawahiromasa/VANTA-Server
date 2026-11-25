package main

import (
	"log"
	"net/http"
	"os"

	"www/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ws", handlers.HandleConnections)

	log.Println("Server started on port", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
