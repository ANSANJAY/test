package main

import (
    "encoding/csv"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/go-ldap/ldap/v3"
)

const (
    BindUsername   = "" // Provide bind username
    BindPassword   = "" // Provide bind password
    FQDN           = "" // Provide LDAP FQDN (e.g., ldap.example.com)
    BaseDN         = "" // Provide base DN (e.g., dc=example,dc=com)
    MaxLevels      = 4  // Set max levels to 4
    OutputFile     = "output.csv"
    ProcessedFile  = "processed.csv" // Output file for successfully processed names
    ErrorFile      = "error.csv"     // Output file for names that caused errors
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("Usage: ./ldap_search input.csv")
    }

    csvFile := os.Args[1]

    // Read CSV file containing emails
    emails, err := ReadEmailsFromCSV(csvFile)
    if err != nil {
        log.Fatal(err)
    }

    // Create the output CSV file
    outputFile, err := os.Create(OutputFile)
    if err != nil {
        log.Fatalf("Could not create output CSV file: %v", err)
    }
    defer outputFile.Close()

    // Create CSV writer for the output file
    outputWriter := csv.NewWriter(outputFile)
    defer outputWriter.Flush()

    // Write the header row to the output CSV
    header := []string{"Level 1 Name", "Level 1 Band", "Level 1 Manager Name", "Level 1 Manager Email"}
    for i := 2; i <= MaxLevels; i++ {
        header = append(header,
            fmt.Sprintf("Level %d Name", i),
            fmt.Sprintf("Level %d Band", i),
            fmt.Sprintf("Level %d Manager Name", i),
            fmt.Sprintf("Level %d Manager Email", i))
    }
    outputWriter.Write(header)

    // Create the processed names CSV file
    processedFile, err := os.Create(ProcessedFile)
    if err != nil {
        log.Fatalf("Could not create processed CSV file: %v", err)
    }
    defer processedFile.Close()

    // Create CSV writer for the processed file
    processedWriter := csv.NewWriter(processedFile)
    defer processedWriter.Flush()

    // Write header for processed file
    processedWriter.Write([]string{"Processed Name"})

    // Create the error names CSV file
    errorFile, err := os.Create(ErrorFile)
    if err != nil {
        log.Fatalf("Could not create error CSV file: %v", err)
    }
    defer errorFile.Close()

    // Create CSV writer for the error file
    errorWriter := csv.NewWriter(errorFile)
    defer errorWriter.Flush()

    // Write header for error file
    errorWriter.Write([]string{"Error Name"})

    // TLS Connection
    l, err := ConnectTLS()
    if err != nil {
        log.Fatal(err)
    }
    defer l.Close()

    // Process each email
    for _, email := range emails {
        fmt.Printf("Processing hierarchy for: %s\n", email)

        // Prepare a row for this email to write to CSV
        row := make([]string, MaxLevels*4)

        // Search the hierarchy
        err = SearchHierarchy(l, email, 0, row, outputWriter)
        if err != nil {
            // Log the error name to the error CSV
            log.Println(err)
            errorWriter.Write([]string{email})
            errorWriter.Flush()
        } else {
            // Log successfully processed name to the processed CSV
            processedWriter.Write([]string{email})
            processedWriter.Flush()
        }
    }
}

// SearchHierarchy searches for the user's info and recursively searches for the manager hierarchy, writing results to CSV.
func SearchHierarchy(l *ldap.Conn, email string, level int, row []string, writer *csv.Writer) error {
    if level >= MaxLevels {
        return nil // Stop after 4 levels
    }

    result, err := BindAndSearch(l, email)
    if err != nil {
        return err
    }

    if len(result.Entries) == 0 {
        return fmt.Errorf("No entries found for %s", email)
    }

    // Extract user's information
    entry := result.Entries[0]
    personName := strings.Split(email, "@")[0]
    band := entry.GetAttributeValue("band")
    managerEmail := entry.GetAttributeValue("mgremail")
    managerName := entry.GetAttributeValue("manageremail")

    // Fill the row for this level
    row[level*4] = personName
    row[level*4+1] = band
    row[level*4+2] = managerName
    row[level*4+3] = managerEmail

    // If we've processed the top-level (4th level), write the row to the CSV
    if level == MaxLevels-1 || managerEmail == "" {
        writer.Write(row)
        writer.Flush()
        return nil
    }

    // Recursively process the manager for the next level
    return SearchHierarchy(l, managerEmail, level+1, row, writer)
}

// ReadEmailsFromCSV reads the CSV file and returns a slice of emails.
func ReadEmailsFromCSV(filePath string) ([]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("Error opening CSV file: %s", err)
    }
    defer file.Close()

    csvReader := csv.NewReader(file)
    records, err := csvReader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("Error reading CSV file: %s", err)
    }

    var emails []string
    for _, record := range records {
        if len(record) > 0 {
            emails = append(emails, record[0]) // Assuming the email is in the first column
        }
    }

    return emails, nil
}

// Ldap Connection with TLS
func ConnectTLS() (*ldap.Conn, error) {
    l, err := ldap.DialURL(fmt.Sprintf("ldaps://%s:636", FQDN))
    if err != nil {
        return nil, err
    }
    return l, nil
}

// Bind and Search
func BindAndSearch(l *ldap.Conn, email string) (*ldap.SearchResult, error) {
    // Bind with provided username and password
    err := l.Bind(BindUsername, BindPassword)
    if err != nil {
        return nil, fmt.Errorf("Bind Error: %s", err)
    }

    // Use escaped filter to avoid LDAP injection
    filter := fmt.Sprintf("(mail=%s)", ldap.EscapeFilter(email))

    // Perform LDAP search with a timeout and request all required attributes
    searchReq := ldap.NewSearchRequest(
        BaseDN,
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0, // unlimited results
        10, // set a timeout of 10 seconds
        false,
        filter,
        []string{"band", "manageremail", "mgremail"}, // Request band, manageremail, and mgremail attributes
        nil,
    )
    result, err := l.Search(searchReq)
    if err != nil {
        return nil, fmt.Errorf("Search Error for %s: %s", email, err)
    }

    return result, nil
}