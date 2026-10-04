package main

import (
	"log"
	"strings"
)

func configureLogging(level string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if strings.EqualFold(level, "debug") {
		log.Printf("debug logging enabled")
	}
}
