package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type printServer struct {
	config Config
}

func newPrintServer(config Config) *printServer {
	return &printServer{config: config}
}

func (server *printServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handlePrint)
	return mux
}

func (server *printServer) handlePrint(w http.ResponseWriter, r *http.Request) {
	configCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("invalid JSON: %v", err)
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := validateBody(body); err != nil {
		log.Printf("validation error: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	printerName := strings.TrimSpace(body.Printer)
	if printerName == "" {
		printerName = server.config.PrinterName
	}
	if printerName == "" {
		writeJSONError(w, http.StatusBadRequest, "printer is required")
		return
	}

	if err := print(body.Operations, printerName); err != nil {
		log.Printf("print error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "print failed")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func configCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONResponse(w, status, map[string]string{"error": message})
}

func writeJSONResponse(w http.ResponseWriter, status int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
