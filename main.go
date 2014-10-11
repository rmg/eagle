package main

import (
	"github.com/rmg/eagle/server"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/metrics", server.MetricsHandler)
	err := http.ListenAndServe(portOrDefault("8000"), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Get the Port from the environment so we can run on Heroku
func portOrDefault(dflt string) string {
	port := os.Getenv("PORT")

	// Set a default port if there is nothing in the environment
	if port == "" {
		port = dflt
	}

	return ":" + port
}
