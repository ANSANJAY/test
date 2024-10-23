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
        fmt.Printf("Searching for: %s\n", email)
        result, err := BindAndSearch(l, email)
        if err != nil {
            log.Println(err)
            continue
        }

        // Print all required attributes if the entry exists
        for _, entry := range result.Entries {
            // Extract the person's name from the email for personalized output
            personName := strings.Split(email, "@")[0] // Get the part before @ as the name
            fmt.Printf("%s's Band: %s\n", personName, entry.GetAttributeValue("band"))
            fmt.Printf("%s's Manager Email: %s\n", personName, entry.GetAttributeValue("manageremail"))
            fmt.Printf("%s's Email: %s\n", personName, entry.GetAttributeValue("mail"))
        }
    }
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
        []string{"band", "manageremail", "mail"}, // Request band, manageremail, and mail attributes
        nil,
    )
    result, err := l.Search(searchReq)
    if err != nil {
        return nil, fmt.Errorf("Search Error for %s: %s", email, err)
    }

    if len(result.Entries) > 0 {
        return result, nil
    } else {
        return nil, fmt.Errorf("No entries found for %s", email)
    }
}