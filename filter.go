package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	// Open wholebatch.csv
	wholebatchFile, err := os.Open("wholebatch.csv")
	if err != nil {
		fmt.Println("Error opening wholebatch.csv:", err)
		return
	}
	defer wholebatchFile.Close()

	// Open maven.csv
	mavenFile, err := os.Open("maven.csv")
	if err != nil {
		fmt.Println("Error opening maven.csv:", err)
		return
	}
	defer mavenFile.Close()

	// Read wholebatch.csv
	wholebatchReader := csv.NewReader(wholebatchFile)
	wholebatchRecords, err := wholebatchReader.ReadAll()
	if err != nil {
		fmt.Println("Error reading wholebatch.csv:", err)
		return
	}

	// Read maven.csv
	mavenReader := csv.NewReader(mavenFile)
	mavenRecords, err := mavenReader.ReadAll()
	if err != nil {
		fmt.Println("Error reading maven.csv:", err)
		return
	}

	// Create a set of Maven project names
	mavenSet := make(map[string]bool)
	for _, mavenRow := range mavenRecords {
		if len(mavenRow) > 0 {
			mavenSet[mavenRow[0]] = true
		}
	}

	// Open filtered_maven.csv for writing
	outputFile, err := os.Create("filtered_maven.csv")
	if err != nil {
		fmt.Println("Error creating filtered_maven.csv:", err)
		return
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write filtered records to the new CSV file
	for _, record := range wholebatchRecords {
		if len(record) > 0 {
			projectName := record[0]
			if _, exists := mavenSet[projectName]; exists {
				if err := writer.Write(record); err != nil {
					fmt.Println("Error writing to filtered_maven.csv:", err)
					return
				}
			}
		}
	}

	fmt.Println("Filtered Maven projects have been written to filtered_maven.csv")
}