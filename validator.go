package main

import (
	"fmt"
	"strconv"
	"strings"
)

func validateBody(body Body) error {
	if len(body.Operations) == 0 {
		return fmt.Errorf("operations are required")
	}

	for _, operation := range body.Operations {
		if strings.TrimSpace(operation.Action) == "" {
			return fmt.Errorf("operation action is required")
		}

		switch operation.Action {
		case "fontSize":
			if strings.TrimSpace(operation.Data) == "" {
				return fmt.Errorf("fontSize data is required")
			}
		case "alignment":
			if operation.Data != "L" && operation.Data != "C" && operation.Data != "R" {
				return fmt.Errorf("alignment must be L, C or R")
			}
		case "text", "enter", "cut", "openCashDrawer", "line", "separator":
		case "boldText":
			if _, err := strconv.Atoi(operation.Data); err != nil {
				return fmt.Errorf("boldText data must be numeric")
			}
		case "feed":
			if _, err := strconv.Atoi(operation.Data); err != nil {
				return fmt.Errorf("feed data must be numeric")
			}
		default:
			return fmt.Errorf("unsupported action: %s", operation.Action)
		}
	}

	return nil
}
