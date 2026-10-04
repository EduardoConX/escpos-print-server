package main

import (
	"log"
	"net/http"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	configureLogging(config.LogLevel)
	server := newPrintServer(config)
	address := ":" + config.Port

	log.Printf("starting server on %s", address)
	if err := http.ListenAndServe(address, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
