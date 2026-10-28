package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host                string `json:"host"`
	Port                int    `json:"port"`
	PrinterName         string `json:"printer_name"`
	LogLevel            string `json:"log_level"`
	APIKey              string `json:"api_key"`
	AllowedOrigins      string `json:"allowed_origins"`
	ReadTimeoutSeconds  int    `json:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `json:"write_timeout_seconds"`
}

func (config Config) address() string {
	return net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
}

func loadConfig() (Config, error) {
	config := Config{
		Host:                "0.0.0.0",
		Port:                8080,
		PrinterName:         "XP-58",
		LogLevel:            "info",
		AllowedOrigins:      "*",
		ReadTimeoutSeconds:  15,
		WriteTimeoutSeconds: 60,
	}

	configPath := getEnv("CONFIG_FILE", "config.json")
	if file, err := os.Open(configPath); err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&config); err != nil {
			return Config{}, fmt.Errorf("read config file %q: %w", configPath, err)
		}
	} else if !os.IsNotExist(err) || os.Getenv("CONFIG_FILE") != "" {
		return Config{}, fmt.Errorf("open config file %q: %w", configPath, err)
	}

	if value := getEnv("HOST", ""); value != "" {
		config.Host = value
	}
	if value := getEnv("PORT", ""); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
		}
		config.Port = port
	}
	if value := getEnv("PRINTER_NAME", ""); value != "" {
		config.PrinterName = value
	}
	if value := getEnv("LOG_LEVEL", ""); value != "" {
		config.LogLevel = value
	}
	if value := getEnv("API_KEY", ""); value != "" {
		config.APIKey = value
	}
	if value := getEnv("ALLOWED_ORIGINS", ""); value != "" {
		config.AllowedOrigins = value
	}
	if value := getEnv("READ_TIMEOUT_SECONDS", ""); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("READ_TIMEOUT_SECONDS must be a positive number")
		}
		config.ReadTimeoutSeconds = seconds
	}
	if value := getEnv("WRITE_TIMEOUT_SECONDS", ""); value != "" {
		seconds, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("WRITE_TIMEOUT_SECONDS must be a positive number")
		}
		config.WriteTimeoutSeconds = seconds
	}
	config.Host = strings.TrimSpace(config.Host)
	config.PrinterName = strings.TrimSpace(config.PrinterName)
	config.LogLevel = strings.ToLower(strings.TrimSpace(config.LogLevel))
	if config.Host == "" {
		return Config{}, fmt.Errorf("HOST must not be empty")
	}
	if config.Port < 1 || config.Port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535")
	}
	if config.PrinterName == "" {
		return Config{}, fmt.Errorf("PRINTER_NAME must not be empty")
	}
	if config.LogLevel != "debug" && config.LogLevel != "info" && config.LogLevel != "warn" && config.LogLevel != "error" {
		return Config{}, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error")
	}
	if config.ReadTimeoutSeconds < 1 || config.WriteTimeoutSeconds < 1 {
		return Config{}, fmt.Errorf("HTTP timeouts must be positive")
	}
	config.AllowedOrigins = strings.TrimSpace(config.AllowedOrigins)
	if config.AllowedOrigins == "" {
		config.AllowedOrigins = "*"
	}
	return config, nil
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
