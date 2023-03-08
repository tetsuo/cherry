package cherry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/onur1/cherry"
	"github.com/onur1/middleware"
	"github.com/stretchr/testify/assert"
)

type entry struct {
	ID string `json:"id"`
}

func (e *entry) Validate() error {
	return validation.ValidateStruct(
		e,
		validation.Field(&e.ID, validation.Length(2, 8)),
	)
}

type testClient struct {
	http.Client
	w   *httptest.ResponseRecorder
	mux *http.ServeMux
}

func (c *testClient) Do(req *http.Request) (*http.Response, error) {
	c.mux.ServeHTTP(c.w, req)
	response := c.w.Result()
	response.Request = req
	return response, nil
}

var (
	echoMiddleware = middleware.ApSecond(
		middleware.ApSecond(
			middleware.Chain(
				middleware.DecodeHeader("baz"),
				func(n string) middleware.Middleware[any] {
					return middleware.Header("foo", "bar+"+n)
				}),
			middleware.Status(201),
		),
		middleware.Chain(
			middleware.DecodeBody[entry],
			middleware.JSON[*entry],
		),
	)
)

func TestCherry(t *testing.T) {
	testCases := []struct {
		desc       string
		request    *cherry.Request[entry]
		middleware middleware.Middleware[any]
		assert     func(*testing.T, *entry, *http.Response)
		assertErr  func(*testing.T, *entry, *http.Response, error)
		expected   entry
	}{
		{
			desc:       "Post",
			middleware: echoMiddleware,
			request: cherry.Post[entry](
				"/",
				&entry{ID: "qux"},
				map[string]string{"baz": "quux+42"},
			),
			assert: func(t *testing.T, expected *entry, res *http.Response) {
				assert.Equal(t, res.Header.Get("foo"), "bar+quux+42")
				assert.Equal(t, res.StatusCode, 201)
				assert.Equal(t, entry{ID: "qux"}, *expected)
			},
		},
		{
			desc:       "BadURL",
			middleware: middleware.Status(404),
			request:    cherry.Get[entry]("/cherry", nil),
			assertErr: func(t *testing.T, entry *entry, resp *http.Response, err error) {
				assert.Equal(t, "bad url", err.Error())
				assert.ErrorIs(t, err, cherry.ErrBadURL)
				assert.Equal(t, "/cherry", resp.Request.URL.String())
				assert.Nil(t, entry)
				assert.NotNil(t, resp)
			},
		},
		{
			desc:       "BadStatus",
			middleware: middleware.Status(400),
			request:    cherry.Get[entry]("/", nil),
			assertErr: func(t *testing.T, entry *entry, resp *http.Response, err error) {
				assert.Equal(t, "bad status", err.Error())
				assert.ErrorIs(t, err, cherry.ErrBadStatus)
				assert.Equal(t, 400, resp.StatusCode)
				assert.Nil(t, entry)
				assert.NotNil(t, resp)
			},
		},
		{
			desc:       "BadPayload",
			middleware: middleware.PlainText("hi"),
			request:    cherry.Get[entry]("/", nil),
			assertErr: func(t *testing.T, entry *entry, resp *http.Response, err error) {
				assert.Equal(t, "invalid character 'h' looking for beginning of value", err.Error())
				assert.Nil(t, entry)
				assert.NotNil(t, resp)
			},
		},
		{
			desc:       "ValidationError",
			middleware: middleware.JSON(&entry{ID: "q"}),
			request:    cherry.Get[entry]("/", nil),
			assertErr: func(t *testing.T, entry *entry, resp *http.Response, err error) {
				assert.Equal(t, "id: the length must be between 2 and 8.", err.Error())
				assert.NotNil(t, entry)
				assert.NotNil(t, resp)
				_, ok := err.(validation.Errors)
				assert.True(t, ok)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.Handle("/", middleware.ToHandlerFunc(tC.middleware, func(err error, c *middleware.Connection) {
				t.Fatal(err.Error())
			}))

			resp, a, err := cherry.Send(&testClient{
				w:   httptest.NewRecorder(),
				mux: mux,
			}, tC.request)
			if tC.assertErr == nil {
				assert.NoError(t, err)
				tC.assert(t, a, resp)
			} else {
				assert.Error(t, err)
				tC.assertErr(t, a, resp, err)
				return
			}
		})
	}
}
