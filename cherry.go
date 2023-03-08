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

type NetworkError = net.Error

type ValidationError = validation.Errors

var (
	ErrTimeout   = errors.New("timeout")
	ErrBadURL    = errors.New("bad url")
	ErrBadStatus = errors.New("bad status")
)

func toRequest[A any](r *Request[A]) (req *http.Request, err error) {
	var body io.Reader
	if r.Body != nil && !(r.Method == "GET" || r.Method == "OPTIONS") {
		var buf []byte
		buf, err = json.Marshal(r.Body)
		if err != nil {
			return
		}
		body = bytes.NewBuffer(buf)
	}
	req, err = http.NewRequest(r.Method, r.URL, body)
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

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

var todo = context.TODO()

func Send[A any](c Client, r *Request[A]) (resp *http.Response, a *A, e error) {
	return SendWithContext(todo, c, r)
}

func SendWithContext[A any](ctx context.Context, c Client, r *Request[A]) (resp *http.Response, a *A, e error) {
	var (
		req *http.Request
		err error
	)
	if req, err = toRequest(r); err != nil {
		e = fmt.Errorf("request: %w", err)
		return
	}
	if ctx != todo {
		req = req.WithContext(ctx)
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

type Request[A any] struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    any
}

func Get[A any](url string, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "GET",
		Headers: headers,
		URL:     url,
	}
}

func Post[A, I any](url string, body *I, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "POST",
		Headers: headers,
		URL:     url,
		Body:    body,
	}
}

func Put[A, I any](url string, body *I, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "PUT",
		Headers: headers,
		URL:     url,
		Body:    body,
	}
}

func Patch[A, I any](url string, body *I, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "PATCH",
		Headers: headers,
		URL:     url,
		Body:    body,
	}
}

func Head[A any](url string, headers map[string]string) *Request[A] {
	return &Request[A]{
		Method:  "PATCH",
		Headers: headers,
		URL:     url,
	}
}
