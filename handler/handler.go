package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"fntv-proxy/logger"
	"fntv-proxy/model"
	"fntv-proxy/store"
)

// HandleProxyInfo handles proxy info setting request
func HandleProxyInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON data from request body
	var requestData model.ProxyInfoRequest

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.URL == "" {
		http.Error(w, "url parameter cannot be empty", http.StatusBadRequest)
		return
	}

	// Store proxy info to global variable
	store.SetProxyInfo(requestData.URL, requestData.Cookie)

	logger.StdoutLogger.Printf("Updated proxy info: URL=%s, Cookie=%s", requestData.URL, requestData.Cookie)

	w.Header().Set("Content-Type", "application/json")
	response := model.CommonResponse{
		Code:    0,
		Status:  "success",
		Message: "Proxy info updated",
		Data:    true,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.StdoutLogger.Printf("Failed to serialize response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleProxyGet handles get proxy info request
func HandleProxyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	info := store.GetProxyInfo()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		logger.StdoutLogger.Printf("Failed to serialize response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleVLCRequest handles VLC requests
func HandleVLCRequest(w http.ResponseWriter, r *http.Request) {
	// Read proxy info
	info := store.GetProxyInfo()
	targetURL := info.URL
	cookie := info.Cookie

	if targetURL == "" {
		http.Error(w, "Proxy info not set, please call /proxyInfo interface first", http.StatusBadRequest)
		return
	}

	requestPath := r.URL.Path
	targetFullURL, err := buildTargetURL(targetURL, requestPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to build target URL: %v", err), http.StatusBadRequest)
		return
	}

	proxyReq, err := http.NewRequest(r.Method, targetFullURL.String(), r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create proxy request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy original request headers
	for name, values := range r.Header {
		if strings.EqualFold(name, "Host") {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Add Cookie
	if cookie != "" {
		proxyReq.Header.Set("Cookie", cookie)
	}

	if strings.Contains(requestPath, "/v/api/v1/media/range/") {
		proxyReq.Header.Set("Range", "bytes=0-")
	}

	// Log request info
	logger.StdoutLogger.Printf("Proxy request: %s %s", r.Method, targetFullURL.String())

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send request: %v", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.StdoutLogger.Printf("Failed to close response body: %v", err)
		}
	}()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Failed to copy response body: %v", err)
	}
}

// buildTargetURL builds the target URL
func buildTargetURL(baseURL, path string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	reference, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	targetURL := u.ResolveReference(reference)
	return targetURL, nil
}
