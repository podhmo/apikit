package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRouter(t *testing.T) {
	r := chi.NewRouter()
	Mount(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// setup global state
	articles = map[int64]*Article{
		1: &Article{
			ID:    1,
			Title: "foo",
		},
	}

	newRequest := func(path string, body io.Reader) *http.Request {
		req, err := http.NewRequest("POST", ts.URL+path, body)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	cases := []struct {
		msg  string
		path string
		body io.Reader
		code int
	}{
		{
			msg:  "invalid path params (not int)",
			path: "/articles/xxxx/comments",
			code: 404,
		},
		{
			msg:  "invalid path params (not found)",
			path: "/articles/0/comments",
			code: 400, // xxx
		},
		{
			msg:  "invalid body (body is nil)",
			path: "/articles/1/comments",
			code: 400,
		},
		{
			msg:  "invalid body (required field is missing)",
			path: "/articles/1/comments",
			code: 422,
			body: strings.NewReader(`{}`),
		},
		{
			msg:  "invalid body (field is zero value)",
			path: "/articles/1/comments",
			code: 422,
			body: strings.NewReader(`{"Text": ""}`),
		},
		{
			msg:  "invalid body (field type is not string)",
			path: "/articles/1/comments",
			code: 400,
			body: strings.NewReader(`{"Text": 1000}`),
		},
		{
			msg:  "ok",
			path: "/articles/1/comments",
			code: 200,
			body: strings.NewReader(`{"Text": "hello"}`),
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			res, err := http.DefaultClient.Do(newRequest(c.path, c.body))
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if want, got := c.code, res.StatusCode; want != got {
				t.Errorf("want code is %v but got %v", want, got)
			}
		})
	}
}
