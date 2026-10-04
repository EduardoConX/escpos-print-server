package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port        string
	PrinterName string
	LogLevel    string
}

func loadConfig() (Config, error) {
	config := Config{
		Port:        getEnv("PORT", "8080"),
		PrinterName: getEnv("PRINTER_NAME", "XP-58"),
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
