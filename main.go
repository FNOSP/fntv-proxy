package main

import (
	"fmt"
	"fntv-proxy/config"
	"fntv-proxy/handler"
	"fntv-proxy/logger"
	"log"
	"net/http"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load config file
	conf, err := config.LoadConfig("config.ini")
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// Start HTTP server
	port := conf.Port
	if port == 0 {
		port = 1999 // Default port
	}

	mux := http.NewServeMux()

	// Setup routes
	mux.HandleFunc("/proxy/info", handler.HandleProxyInfo)
	mux.HandleFunc("/proxyGet", handler.HandleProxyGet)
	mux.HandleFunc("/", handler.HandleVLCRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	logger.StdoutLogger.Printf("VLC Proxy service started, listening on port: %d", port)
	log.Fatal(server.ListenAndServe())
}
