//go:build windows
// +build windows

package main

import (
	"fmt"
	"io"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	winspoolDLL       = syscall.NewLazyDLL("winspool.drv")
	openPrinterProc   = winspoolDLL.NewProc("OpenPrinterW")
	closePrinterProc  = winspoolDLL.NewProc("ClosePrinter")
	startDocPrinter   = winspoolDLL.NewProc("StartDocPrinterW")
	endDocPrinter     = winspoolDLL.NewProc("EndDocPrinter")
	startPagePrinter  = winspoolDLL.NewProc("StartPagePrinter")
	endPagePrinter    = winspoolDLL.NewProc("EndPagePrinter")
	writePrinterBytes = winspoolDLL.NewProc("WritePrinter")
)

type documentInfo1 struct {
	documentName *uint16
	outputFile   *uint16
	dataType     *uint16
}

func sendToPrinter(printer string, data []byte) error {
	printerName, err := syscall.UTF16PtrFromString(printer)
	if err != nil {
		return err
	}
	documentName, err := syscall.UTF16PtrFromString("ESC/POS ticket")
	if err != nil {
		return err
	}
	rawDataType, err := syscall.UTF16PtrFromString("RAW")
	if err != nil {
		return err
	}

	var printerHandle syscall.Handle
	result, _, callErr := openPrinterProc.Call(
		uintptr(unsafe.Pointer(printerName)),
		uintptr(unsafe.Pointer(&printerHandle)),
		0,
	)
	if err := winspoolCallError("OpenPrinterW", result, callErr); err != nil {
		return err
	}
	defer closePrinterProc.Call(uintptr(printerHandle))

	documentInfo := documentInfo1{
		documentName: documentName,
		dataType:     rawDataType,
	}
	jobID, _, callErr := startDocPrinter.Call(
		uintptr(printerHandle),
		1,
		uintptr(unsafe.Pointer(&documentInfo)),
	)
	if err := winspoolCallError("StartDocPrinterW", jobID, callErr); err != nil {
		return err
	}
	documentStarted := true
	defer func() {
		if documentStarted {
			endDocPrinter.Call(uintptr(printerHandle))
		}
	}()

	result, _, callErr = startPagePrinter.Call(uintptr(printerHandle))
	if err := winspoolCallError("StartPagePrinter", result, callErr); err != nil {
		return err
	}
	pageStarted := true
	defer func() {
		if pageStarted {
			endPagePrinter.Call(uintptr(printerHandle))
		}
	}()

	var bytesWritten uint32
	result, _, callErr = writePrinterBytes.Call(
		uintptr(printerHandle),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	runtime.KeepAlive(data)
	if err := winspoolCallError("WritePrinter", result, callErr); err != nil {
		return err
	}
	if uint64(bytesWritten) != uint64(len(data)) {
		return io.ErrShortWrite
	}

	result, _, callErr = endPagePrinter.Call(uintptr(printerHandle))
	if err := winspoolCallError("EndPagePrinter", result, callErr); err != nil {
		return err
	}
	pageStarted = false

	result, _, callErr = endDocPrinter.Call(uintptr(printerHandle))
	if err := winspoolCallError("EndDocPrinter", result, callErr); err != nil {
		return err
	}
	documentStarted = false
	return nil
}

func winspoolCallError(operation string, result uintptr, callErr error) error {
	if result != 0 {
		return nil
	}
	if errno, ok := callErr.(syscall.Errno); ok && errno != 0 {
		return fmt.Errorf("%s: %w", operation, errno)
	}
	return fmt.Errorf("%s failed", operation)
}
