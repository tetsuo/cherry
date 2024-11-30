package cherry_test

import (
	"fmt"
	"log"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/tetsuo/cherry"
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

func (e *Index) Print() {
	for _, entry := range e.Entries {
		fmt.Printf(
			"%s [https://ogu.nz/%s.html]\n%s\n\n",
			entry.Title, entry.Slug, entry.Description,
		)
	}
}

func ExampleSend() {
	req := cherry.Get[Index]("https://ogu.nz/index.json", nil)

	resp, index, err := cherry.Send(http.DefaultClient, req)
	if err != nil {
		log.Fatalf("send failed: %v", err)
	}

	fmt.Println(resp.StatusCode)

	fmt.Println(len(index.Entries) > 0)

	// index.Print()

	// Output:
	// 200
	// true
}

func ExampleErrBadURL() {
	req := cherry.Get[Index]("https://ogu.nz/undefined", nil)

	resp, _, err := cherry.Send(http.DefaultClient, req)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(err == cherry.ErrBadURL)
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// bad url
	// true
	// 404
}

type unknownEntryType struct {
	ID string `json:"id"`
}

func (e *unknownEntryType) Validate() error {
	return validation.ValidateStruct(
		e,
		validation.Field(&e.ID, validation.Required, validation.Length(2, 8)),
	)
}

func ExampleValidationError() {
	req := cherry.Get[unknownEntryType]("https://ogu.nz/index.json", nil)

	resp, _, err := cherry.Send(http.DefaultClient, req)
	if err != nil {
		validationErrors, ok := err.(validation.Errors)
		fmt.Println(ok)
		fmt.Println(validationErrors)
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// true
	// id: cannot be blank.
	// 200
}
