package main

import (
	"crypto/subtle"
	"encoding/json"
	"io"
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
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/status", server.handleStatus)
	mux.HandleFunc("/test", server.handleTestPrint)
	mux.HandleFunc("/", server.handlePrint)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.configCORS(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (server *printServer) handlePrint(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !server.authorized(r) {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body Body
	if err := decodeRequest(w, r, &body); err != nil {
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

func (server *printServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (server *printServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !server.authorized(r) {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":         "ok",
		"printer_name":   server.config.PrinterName,
		"printer_status": "not_checked",
	})
}

func (server *printServer) handleTestPrint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !server.authorized(r) {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := print([]Operation{
		{Action: "alignment", Data: "C"},
		{Action: "text", Data: "ESC/POS TEST PRINT"},
		{Action: "enter"},
		{Action: "text", Data: "Connection successful"},
		{Action: "feed", Data: "2"},
	}, server.config.PrinterName); err != nil {
		log.Printf("test print error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "test print failed")
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeRequest(w http.ResponseWriter, r *http.Request, destination interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}
	return http.ErrBodyNotAllowed
}

func (server *printServer) authorized(r *http.Request) bool {
	if server.config.APIKey == "" {
		return true
	}
	const prefix = "Bearer "
	provided := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), prefix))
	return strings.HasPrefix(r.Header.Get("Authorization"), prefix) &&
		subtle.ConstantTimeCompare([]byte(provided), []byte(server.config.APIKey)) == 1
}

func (server *printServer) configCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	for _, allowed := range strings.Split(server.config.AllowedOrigins, ",") {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" || (origin != "" && origin == allowed) {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			if allowed != "*" {
				w.Header().Add("Vary", "Origin")
			}
			break
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
