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

func main() {
	http.HandleFunc("/", handler)

	fmt.Printf("Starting server...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	configCORS(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}

	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var b Body
		err := decoder.Decode(&b)
		if err != nil {
			panic(err)
		}
		print(b.Operations, b.Printer)
	}
}

func configCORS(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func print(operations []Operation, printer string) {
	fileName := "printFile"
	//Create file to print
	f, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}

	//Close file after creation
	defer f.Close()

	w := bufio.NewWriter(f)
	w.Write(startPrinter())

	for _, operation := range operations {
		w.Write(operationsHandler(operation))
	}

	w.Flush()
	copyToPrinter(fileName, printer)
}

func operationsHandler(operation Operation) []byte {
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
		return enter()
	}

	return []byte("")
}

func startPrinter() []byte {
	return []byte("\x1B@")
}

func fontSize(datos string) []byte {
	values := strings.Split(datos, ",")
	width, err := strconv.Atoi(values[0])
	if err != nil {
		panic(err)
	}

	height, err := strconv.Atoi(values[1])
	if err != nil {
		panic(err)
	}
	return []byte(fmt.Sprintf("\x1D!%c", ((width-1)<<4)|(height-1)))
}

func alignment(alignment string) []byte {
	realAlignment := 0
	switch alignment {
	case "L":
		realAlignment = 0
	case "C":
		realAlignment = 1
	case "R":
		realAlignment = 2
	}
	return []byte(fmt.Sprintf("\x1Ba%c", realAlignment))
}

func text(text string) []byte {
	//Decodes accents and ñ
	c, e := charmap.CodePage850.NewEncoder().String(text)

	if e != nil {
		log.Fatal(e)
	}

	return []byte(c)
}

func boldText(enable string) []byte {
	isEnable, err := strconv.Atoi(enable)
	if err != nil {
		panic(err)
	}
	return []byte(fmt.Sprintf("\x1B\x45%c", isEnable))
}

func feed(nLines string) []byte {
	n, err := strconv.Atoi(nLines)
	if err != nil {
		panic(err)
	}
	return []byte(fmt.Sprintf("\x1Bd%c", n))
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
	fd2, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false, err
	}
	defer fd2.Close()
	_, e := io.Copy(fd2, fd1)
	if e != nil {
		return false, e
	}
	return true, nil
}
