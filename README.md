# cherry

simple go http client.

## Example

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/onur1/cherry"
)

type Entry struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
}

func (e *Entry) Validate() error {
	return validation.ValidateStruct(
		e,
		validation.Field(&e.Title, validation.Required),
		validation.Field(&e.Description, validation.Required),
		validation.Field(&e.Slug, validation.Required),
	)
}

type Index struct {
	Entries []*Entry `json:"entries"`
}

func main() {
	req := cherry.Get[Index]("https://ogu.nz/index.json", nil)

	resp, index, err := cherry.Send(http.DefaultClient, req)
	if err != nil {
		log.Fatalf("send failed: %v", err)
	}

	fmt.Println(resp.StatusCode)

	fmt.Println(len(index.Entries) > 0)

	for _, v := range index.Entries {
		fmt.Println(v.Title)
	}
}
```

Output:

```
200
true
flixbox
binproto
tango
couchilla
This site
```
