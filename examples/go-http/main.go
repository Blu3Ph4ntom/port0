package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
			<h1>Go HTTP Server</h1>
			<p>Port: %s</p>
			<p>Host: %s</p>
			<p>Path: %s</p>
		`, port, r.Host, r.URL.Path)
	})

	fmt.Printf("Go HTTP server listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
