package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func main() {
	// Open match_link.csv
	matchLinkFile, err := os.Open("match_link.csv")
	if err != nil {
		log.Fatalf("Failed to open match_link.csv: %v", err)
	}
	defer matchLinkFile.Close()

	// Read match_link.csv
	matchLinkReader := csv.NewReader(matchLinkFile)
	matchLinkData, err := matchLinkReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read match_link.csv: %v", err)
	}

	// Open max_output.csv
	maxOutputFile, err := os.Open("max_output.csv")
	if err != nil {
		log.Fatalf("Failed to open max_output.csv: %v", err)
	}
	defer maxOutputFile.Close()

	// Read max_output.csv
	maxOutputReader := csv.NewReader(maxOutputFile)
	maxOutputData, err := maxOutputReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read max_output.csv: %v", err)
	}

	// Create a map from max_output.csv (name -> id)
	nameToID := make(map[string]string)
	for _, row := range maxOutputData {
		if len(row) < 2 {
			continue // Skip invalid rows
		}
		nameToID[row[0]] = row[1]
	}

	// Prepare a new CSV file to store matched results
	outputFile, err := os.Create("matched_output.csv")
	if err != nil {
		log.Fatalf("Failed to create matched_output.csv: %v", err)
	}
	defer outputFile.Close()

	outputWriter := csv.NewWriter(outputFile)
	defer outputWriter.Flush()

	// Write matched rows to the new CSV
	for _, row := range matchLinkData {
		if len(row) < 1 {
			continue // Skip invalid rows
		}
		name := row[0]
		if id, found := nameToID[name]; found {
			err := outputWriter.Write([]string{name, id})
			if err != nil {
				log.Fatalf("Failed to write to matched_output.csv: %v", err)
			}
		}
	}

	fmt.Println("Matching complete. Results written to matched_output.csv")
}
