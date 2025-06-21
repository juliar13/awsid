package main

import (
	"fmt"
)

// ValidFormats defines the list of valid format options
var ValidFormats = []string{"json", "table", "csv", "default"}

// ResolveFormatFlags resolves format flag conflicts and returns the final format
// Priority: --format flag takes priority over individual flags (--json, --table, --csv)
func ResolveFormatFlags(formatFlag string, jsonFlag, tableFlag, csvFlag bool) (string, error) {
	// Count individual format flags
	individualFlags := 0
	var individualFormat string
	
	if jsonFlag {
		individualFlags++
		individualFormat = "json"
	}
	if tableFlag {
		individualFlags++
		individualFormat = "table"
	}
	if csvFlag {
		individualFlags++
		individualFormat = "csv"
	}
	
	// Case 1: --format flag is specified
	if formatFlag != "" {
		// Validate format flag value
		if !isValidFormat(formatFlag) {
			return "", fmt.Errorf("invalid format '%s'. Valid formats are: %v", formatFlag, ValidFormats)
		}
		
		// --format takes priority, ignore individual flags
		return formatFlag, nil
	}
	
	// Case 2: Multiple individual flags specified (error)
	if individualFlags > 1 {
		return "", fmt.Errorf("multiple format flags specified. Please use only one format flag")
	}
	
	// Case 3: Single individual flag specified
	if individualFlags == 1 {
		return individualFormat, nil
	}
	
	// Case 4: No format flags specified
	return "default", nil
}

// isValidFormat checks if the given format is valid
func isValidFormat(format string) bool {
	for _, valid := range ValidFormats {
		if format == valid {
			return true
		}
	}
	return false
}

// ParseFormatFlags is a helper function to extract format information from command flags
// This would typically be called from the main command handler
func ParseFormatFlags(cmd interface{}) (formatFlag string, jsonFlag, tableFlag, csvFlag bool, err error) {
	// This is a placeholder for actual flag parsing
	// In real implementation, this would extract flags from cobra.Command
	// For now, return empty values as this will be integrated with main.go
	return "", false, false, false, nil
}