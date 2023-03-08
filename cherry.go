package cherry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
)

// ValidationError is an alias to validation.Errors from the ozzo-validation
// library.
type ValidationError = validation.Errors

var (
	// ErrTimeout represents a timeout error.
	ErrTimeout = errors.New("timeout")
	// ErrBadURL represents a 404 not found response.
	ErrBadURL = errors.New("bad url")
	// ErrBadStatus represents a response with a status code which is not 2xx.
	ErrBadStatus = errors.New("bad status")
	// ErrBadRequest wraps any error that might happen while converting a Request
	// into an http.Request.
	ErrBadRequest = errors.New("bad request")
)

func toRequestWithContext[A any](ctx context.Context, r *Request[A]) (req *http.Request, err error) {
	var body io.Reader
	if r.Body != nil && !(r.Method == "GET" || r.Method == "OPTIONS") {
		var buf []byte
		buf, err = json.Marshal(r.Body)
		if err != nil {
			return
		}
		body = bytes.NewBuffer(buf)
	}
	req, err = http.NewRequestWithContext(ctx, r.Method, r.URL, body)
	if err != nil {
		return
	}
	for key, value := range r.Headers {
		req.Header.Add(key, value)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "cherry/1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return
}

// A Client manages the HTTP connection.
type Client interface {
	// Do sends an HTTP request and returns an HTTP response, following policy
	// (such as redirects, cookies, auth) as configured on the client.
	Do(*http.Request) (*http.Response, error)
}

// Send creates and sends a new http.Request, returning an HTTP response and
// a pointer to a value of type A along with an error if any encountered.
func Send[A any](c Client, r *Request[A]) (resp *http.Response, a *A, e error) {
	return SendWithContext(context.Background(), c, r)
}

// SendWithContext creates and sends a new context-aware http.Request, returning
// an HTTP response and a pointer to a value of type A along with an error if
// any encountered.
func SendWithContext[A any](ctx context.Context, c Client, r *Request[A]) (resp *http.Response, a *A, e error) {
	var (
		req *http.Request
		err error
	)
	if req, err = toRequestWithContext(ctx, r); err != nil {
		e = fmt.Errorf("%w: %w", ErrBadRequest, err)
		return
	}
	if resp, err = c.Do(req); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			e = ErrTimeout
			return
		}
		e = err
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 404 {
			e = ErrBadURL
			return
		}
		e = ErrBadStatus
		return
	}
	a = new(A)
	if err = json.NewDecoder(resp.Body).Decode(a); err != nil {
		a = nil
		e = err
		return
	}
	if validationErrors := validation.Validate(a); validationErrors != nil {
		e = validationErrors
	}
	return
}

// A Request represents an HTTP request to be sent by a client.
type Request[A any] struct {
	// URL specifies the URI being requested.
	URL string
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	Method string
	// Headers contain the request header fields to be sent.
	Headers map[string]string
	// Body is request's body.
	Body any
}

// Get creates a new GET a return value of type A.
func Get[A any](url string, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "GET",
		Headers: headers,
		URL:     url,
	}
}

// Post creates a new POST request with the specified payload and with
// a return value of type A.
func Post[A, I any](url string, body *I, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "POST",
		Headers: headers,
		URL:     url,
		Body:    body,
	}
}

// Put creates a new PUT request with the specified payload and with
// a return value of type A.
func Put[A, I any](url string, body *I, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "PUT",
		Headers: headers,
		URL:     url,
		Body:    body,
	}
}

// Patch creates a new PATCH request with the specified payload and with
// a return value of type A.
func Patch[A, I any](url string, body *I, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "PATCH",
		Headers: headers,
		URL:     url,
		Body:    body,
	}
}
