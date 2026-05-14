package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"converter/internal/handler"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/convert", handler.ConvertHandler)
	mux.HandleFunc("/api/detect", handler.DetectHandler)
	mux.HandleFunc("/api/convert-file", handler.ConvertFileHandler)

	// Static frontend
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))

	log.Printf("Server running at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
