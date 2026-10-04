//go:build !windows
// +build !windows

package main

import (
	"os"
)

func sendToPrinter(destination string, data []byte) error {
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
