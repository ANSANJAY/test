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
    BindUsername = "" // Provide bind username
    BindPassword = "" // Provide bind password
    FQDN         = "" // Provide LDAP FQDN (e.g., ldap.example.com)
    BaseDN       = "" // Provide base DN (e.g., dc=example,dc=com)
    MaxLevels    = 4  // Set max levels to 4
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

    // TLS Connection
    l, err := ConnectTLS()
    if err != nil {
        log.Fatal(err)
    }
    defer l.Close()

    // Bind and search for each email
    for _, email := range emails {
        fmt.Printf("Searching hierarchy for: %s\n", email)
        err = SearchHierarchy(l, email, 0)
        if err != nil {
            log.Println(err)
        }
    }
}

// SearchHierarchy searches for the user's info and recursively searches for the manager hierarchy up to 4 levels.
func SearchHierarchy(l *ldap.Conn, email string, level int) error {
    if level >= MaxLevels {
        return nil // Stop after 4 levels
    }

    indent := strings.Repeat("  ", level) // For indentation to represent hierarchy level

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

    // Print user's information with indentation for hierarchy
    fmt.Printf("%s%s's Band: %s\n", indent, personName, band)
    fmt.Printf("%s%s's Manager Name: %s\n", indent, personName, managerName)
    fmt.Printf("%s%s's Manager Email: %s\n", indent, personName, managerEmail)

    // If manager email is not empty, recursively search for the manager
    if managerEmail != "" {
        fmt.Printf("%sSearching for %s's manager info...\n", indent, personName)
        return SearchHierarchy(l, managerEmail, level+1)
    }

    return nil
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
        []string{"band", "manageremail", "mgremail"}, // Request band, manageremail, and mail attributes
        nil,
    )
    result, err := l.Search(searchReq)
    if err != nil {
        return nil, fmt.Errorf("Search Error for %s: %s", email, err)
    }

    return result, nil
}