# cherry

**cherry** is a simple HTTP client for Go with JSON decoding and request validation. It simplifies HTTP requests and responses by automatically handling JSON data and providing type validation through [ozzo-validation](https://github.com/go-ozzo/ozzo-validation).

## Features

- **Simplified HTTP Requests**: Makes HTTP requests easy with default configurations.
- **Automatic JSON Decoding**: Parses JSON responses directly into structs.
- **Integrated Validation**: Uses `ozzo-validation` for type-safe request and response validation.

## Installation

To install the `cherry` package, use:

```sh
go get github.com/onur1/cherry
```

## Usage

Here's an example demonstrating how to use `cherry` to fetch and validate JSON data from an endpoint:

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    validation "github.com/go-ozzo/ozzo-validation"
    "github.com/onur1/cherry"
)

// Entry represents a single entry in the JSON response
type Entry struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Slug        string `json:"slug"`
}

// Validate checks the required fields for an Entry
func (e *Entry) Validate() error {
    return validation.ValidateStruct(
        e,
        validation.Field(&e.Title, validation.Required),
        validation.Field(&e.Description, validation.Required),
        validation.Field(&e.Slug, validation.Required),
    )
}

// Index represents the structure of the JSON response
type Index struct {
    Entries []*Entry `json:"entries"`
}

func main() {
    // Prepare a GET request with expected response type
    req := cherry.Get[Index]("https://ogu.nz/index.json", nil)

    // Send the request using Cherry and decode JSON into Index
    resp, index, err := cherry.Send(http.DefaultClient, req)
    if err != nil {
        log.Fatalf("send failed: %v", err)
    }

    fmt.Println(resp.StatusCode)            // HTTP status code
    fmt.Println(len(index.Entries) > 0)     // True if entries are present

    // Print titles of all entries
    for _, entry := range index.Entries {
        fmt.Println(entry.Title)
    }
}
```

### Example Output

```
200
true
flixbox
binproto
tango
couchilla
This site
```

## License

MIT License. See [LICENSE](LICENSE) for details.
