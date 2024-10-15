package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func main() {
	// Open sheet_with_car.csv
	carFile, err := os.Open("sheet_with_car.csv")
	if err != nil {
		log.Fatalf("Failed to open sheet_with_car.csv: %v", err)
	}
	defer carFile.Close()

	// Open all_data.csv
	allDataFile, err := os.Open("all_data.csv")
	if err != nil {
		log.Fatalf("Failed to open all_data.csv: %v", err)
	}
	defer allDataFile.Close()

	// Read sheet_with_car.csv
	carReader := csv.NewReader(carFile)
	carData, err := carReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read sheet_with_car.csv: %v", err)
	}

	// Read all_data.csv
	allDataReader := csv.NewReader(allDataFile)
	allData, err := allDataReader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read all_data.csv: %v", err)
	}

	// Create a map to store repo_name -> id_or_error from sheet_with_car
	repoMap := make(map[string]string)
	for _, row := range carData[1:] { // Skip the header row
		if len(row) >= 2 {
			repoMap[row[0]] = row[1] // repo_name -> id_or_error
		}
	}

	// Add a new "central_id" column header to all_data if not present
	header := append(allData[0], "central_id")

	// Prepare the new data with the "central_id" column
	var updatedData [][]string
	updatedData = append(updatedData, header) // Add the header first

	// Loop through all_data rows to find matches and append the "central_id"
	for _, row := range allData[1:] { // Skip the header
		if len(row) > 0 {
			repoName := row[0]
			if centralID, found := repoMap[repoName]; found {
				row = append(row, centralID) // Append the matching id_or_error
			} else {
				row = append(row, "") // No match, append an empty value
			}
		}
		updatedData = append(updatedData, row) // Add the row to updatedData
	}

	// Create a new CSV file to save the updated data
	outputFile, err := os.Create("all_data_updated.csv")
	if err != nil {
		log.Fatalf("Failed to create all_data_updated.csv: %v", err)
	}
	defer outputFile.Close()

	// Write the updated data to the new CSV file
	writer := csv.NewWriter(outputFile)
	err = writer.WriteAll(updatedData)
	if err != nil {
		log.Fatalf("Failed to write to all_data_updated.csv: %v", err)
	}

	fmt.Println("Matching and insertion completed successfully!")
}
