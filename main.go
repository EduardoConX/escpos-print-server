package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/text/encoding/charmap"
)

type Operation struct {
	Action string
	Data   string
}

type Body struct {
	Operations []Operation
	Printer    string
}

var printMu sync.Mutex

func main() {
	configureLogging()
	port := getEnv("PORT", "8080")
	printerPath := getEnv("PRINTER_PATH", "")

	if printerPath != "" {
		log.Printf("printer path configured: %s", printerPath)
	}

	http.HandleFunc("/", handler)

	fmt.Printf("Starting server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func configureLogging() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	level := strings.ToLower(getEnv("LOG_LEVEL", "info"))
	if level == "debug" {
		log.Printf("logging configured at debug level")
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func resolvePrinterDestination(requestPrinter string) string {
	configuredPrinter := strings.TrimSpace(getEnv("PRINTER_PATH", ""))
	if strings.TrimSpace(requestPrinter) != "" {
		return requestPrinter
	}
	return configuredPrinter
}

func handler(w http.ResponseWriter, r *http.Request) {
	configCORS(&w, r)

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

	printerPath := resolvePrinterDestination(body.Printer)
	if strings.TrimSpace(printerPath) == "" {
		writeJSONError(w, http.StatusBadRequest, "printer is required")
		return
	}

	if err := print(body.Operations, printerPath); err != nil {
		log.Printf("print error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "print failed")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSONResponse(w, status, map[string]string{"error": msg})
}

func writeJSONResponse(w http.ResponseWriter, status int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func validateBody(b Body) error {
	if len(b.Operations) == 0 {
		return fmt.Errorf("operations are required")
	}

    for _, op := range b.Operations {
        if strings.TrimSpace(op.Action) == "" {
            return fmt.Errorf("operation action is required")
        }
        switch op.Action {
        case "fontSize":
            if strings.TrimSpace(op.Data) == "" {
                return fmt.Errorf("fontSize data is required")
            }
        case "alignment":
            if op.Data != "L" && op.Data != "C" && op.Data != "R" {
                return fmt.Errorf("alignment must be L, C or R")
            }
        case "text":
            // valid as long as data is present
        case "boldText":
            if _, err := strconv.Atoi(op.Data); err != nil {
                return fmt.Errorf("boldText data must be numeric")
            }
        case "feed":
            if _, err := strconv.Atoi(op.Data); err != nil {
                return fmt.Errorf("feed data must be numeric")
            }
        case "enter":
        default:
            return fmt.Errorf("unsupported action: %s", op.Action)
        }
    }

    return nil
}

func configCORS(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func print(operations []Operation, printer string) error {
	printMu.Lock()
	defer printMu.Unlock()

	tempFile, err := os.CreateTemp("", "escpos-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	writer := bufio.NewWriter(tempFile)
	if _, err := writer.Write(startPrinter()); err != nil {
		return err
	}

	for _, operation := range operations {
		data, err := operationsHandler(operation)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if _, err := copyToPrinter(tempFile.Name(), printer); err != nil {
		return err
	}

	return nil
}

func operationsHandler(operation Operation) ([]byte, error) {
	switch operation.Action {
	case "fontSize":
		return fontSize(operation.Data)
	case "alignment":
		return alignment(operation.Data)
	case "text":
		return text(operation.Data)
	case "boldText":
		return boldText(operation.Data)
	case "feed":
		return feed(operation.Data)
	case "enter":
		return enter(), nil
	default:
		return nil, fmt.Errorf("unsupported action: %s", operation.Action)
	}
}

func startPrinter() []byte {
	return []byte("\x1B@")
}

func fontSize(datos string) ([]byte, error) {
    values := strings.Split(datos, ",")
    if len(values) != 2 {
        return nil, fmt.Errorf("fontSize requires width,height")
    }

    width, err := strconv.Atoi(values[0])
    if err != nil {
        return nil, err
    }

    height, err := strconv.Atoi(values[1])
    if err != nil {
        return nil, err
    }

    return []byte(fmt.Sprintf("\x1D!%c", ((width-1)<<4)|(height-1))), nil
}

func alignment(alignment string) ([]byte, error) {
	realAlignment := 0
	switch alignment {
	case "L":
		realAlignment = 0
	case "C":
		realAlignment = 1
	case "R":
		realAlignment = 2
	default:
		return nil, fmt.Errorf("alignment must be L, C or R")
	}
	return []byte(fmt.Sprintf("\x1Ba%c", realAlignment)), nil
}

func text(value string) ([]byte, error) {
	encoded, err := charmap.CodePage850.NewEncoder().String(value)
	if err != nil {
		return nil, err
	}
	return []byte(encoded), nil
}

func boldText(enable string) ([]byte, error) {
	isEnable, err := strconv.Atoi(enable)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\x1B\x45%c", isEnable)), nil
}

func feed(nLines string) ([]byte, error) {
	n, err := strconv.Atoi(nLines)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\x1Bd%c", n)), nil
}

func enter() []byte {
	return []byte("\n")
}

func copyToPrinter(source, dest string) (bool, error) {
	fd1, err := os.Open(source)
	if err != nil {
		return false, err
	}
	defer fd1.Close()

	fd2, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}
	defer fd2.Close()

	if _, e := io.Copy(fd2, fd1); e != nil {
		return false, e
	}
	return true, nil
}
