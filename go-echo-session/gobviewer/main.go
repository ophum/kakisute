package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"log/slog"
	"strings"
)

func main() {
	// Read base64 encoded string from stdin
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	// Decode the base64 input
	decodedInput, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		log.Fatalf("Failed to decode base64 input: %v", err)
	}

	// Log the decoded input
	fmt.Printf("Decoded input: %s\n", string(decodedInput))

	// Split the decoded data by "|"
	parts := strings.Split(string(decodedInput), "|")
	if len(parts) != 3 {
		log.Fatal("Invalid decoded format. Expected: timestamp|base64data|mac")
	}

	// Extract the base64 encoded data (middle part)
	base64Data := parts[1]

	// Decode base64 string
	decodedData, err := base64.URLEncoding.DecodeString(base64Data)
	if err != nil {
		log.Fatalf("Failed to decode base64: %v", err)
	}

	// Try to decode gob data
	var result map[any]any
	decoder := gob.NewDecoder(bytes.NewReader(decodedData))
	err = decoder.Decode(&result)
	if err != nil {
		slog.Error("failed to decode", "error", err)
		// If gob decoding fails, just print the raw decoded data as JSON
		return
	}

	slog.Info("decoded", "result", result)
}
