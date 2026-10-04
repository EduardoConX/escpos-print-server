package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	PrinterPath string
	LogLevel    string
}

func loadConfig() (Config, error) {
	config := Config{
		Port:        getEnv("PORT", "8080"),
		PrinterPath: getEnv("PRINTER_PATH", ""),
		LogLevel:    strings.ToLower(getEnv("LOG_LEVEL", "info")),
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
	}

	return config, nil
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
