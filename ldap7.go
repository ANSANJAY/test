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
    BindUsername = "name"
    BindPassword = "pwd"
    FQDN         = "ldap"
    BaseDN       = "insert required info"
    MaxLevels    = 4 // Set max levels to 4
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
    managerEmail := entry.GetAttributeValue("manageremail")

    // Print user's information with indentation for hierarchy
    fmt.Printf("%s%s's Band: %s\n", indent, personName, band)
    fmt.Printf("%s%s's Email: %s\n", indent, personName, email)

    // If manager email is not empty, retrieve and print manager's name and email
    if managerEmail != "" {
        managerName, err := GetManagerName(l, managerEmail)
        if err != nil {
            log.Println(err)
        } else {
            fmt.Printf("%s%s's Manager: %s <%s>\n", indent, personName, managerName, managerEmail)
        }

        fmt.Printf("%sSearching for %s's manager info...\n", indent, personName)
        return SearchHierarchy(l, managerEmail, level+1)
    }

    return nil
}

// GetManagerName performs an LDAP search for the manager's email and returns the manager's name.
func GetManagerName(l *ldap.Conn, managerEmail string) (string, error) {
    result, err := BindAndSearch(l, managerEmail)
    if err != nil {
        return "", fmt.Errorf("Error searching for manager %s: %s", managerEmail, err)
    }

    if len(result.Entries) == 0 {
        return "", fmt.Errorf("No entries found for manager %s", managerEmail)
    }

    // Assuming the common name (cn) or displayName contains the manager's name
    managerName := result.Entries[0].GetAttributeValue("cn") // Use "displayName" if that's what your LDAP uses

    if managerName == "" {
        managerName = "Unknown"
    }

    return managerName, nil
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
        []string{"cn", "band", "manageremail", "mail"}, // Request cn (common name), band, manageremail, and mail attributes
        nil,
    )
    result, err := l.Search(searchReq)
    if err != nil {
        return nil, fmt.Errorf("Search Error for %s: %s", email, err)
    }

    return result, nil
}