package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

func print(operations []Operation, printer string) error {
	return printerQueue.run(func() error {
		var output bytes.Buffer
		writer := bufio.NewWriter(&output)
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
		return sendToPrinter(printer, output.Bytes())
	})
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
	case "cut":
		return cut(), nil
	case "openCashDrawer":
		return openCashDrawer(), nil
	case "separator":
		return separator(operation.Data), nil
	case "line":
		return line(), nil
	default:
		return nil, fmt.Errorf("unsupported action: %s", operation.Action)
	}
}

func startPrinter() []byte {
	return []byte("\x1B@")
}

func cut() []byte {
	return []byte("\x1DVA0")
}

func openCashDrawer() []byte {
	return []byte("\x1B\x70\x00\x50\x50")
}

func separator(char string) []byte {
	if char == "" {
		char = "-"
	}
	return []byte(strings.Repeat(char, 42) + "\n")
}

func line() []byte {
	return []byte("\n")
}

func fontSize(data string) ([]byte, error) {
	values := strings.Split(data, ",")
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

func alignment(value string) ([]byte, error) {
	position := 0
	switch value {
	case "L":
		position = 0
	case "C":
		position = 1
	case "R":
		position = 2
	default:
		return nil, fmt.Errorf("alignment must be L, C or R")
	}
	return []byte(fmt.Sprintf("\x1Ba%c", position)), nil
}

func text(value string) ([]byte, error) {
	encoded, err := charmap.CodePage850.NewEncoder().String(value)
	if err != nil {
		return nil, err
	}
	return []byte(encoded), nil
}

func boldText(enable string) ([]byte, error) {
	value, err := strconv.Atoi(enable)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\x1B\x45%c", value)), nil
}

func feed(lines string) ([]byte, error) {
	value, err := strconv.Atoi(lines)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\x1Bd%c", value)), nil
}

func enter() []byte {
	return []byte("\n")
}
