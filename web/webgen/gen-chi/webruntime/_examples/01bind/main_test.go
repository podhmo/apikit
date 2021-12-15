package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func Test(t *testing.T) {
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

		code       int
		resultBody string
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
			body: strings.NewReader(`{"text": ""}`),
		},
		{
			msg:  "invalid body (field type is not string)",
			path: "/articles/1/comments",
			code: 400,
			body: strings.NewReader(`{"text": 1000}`),
		},
		{
			msg:        "ok",
			path:       "/articles/1/comments",
			code:       200,
			body:       strings.NewReader(`{"text": "hello"}`),
			resultBody: `{"author":"someone","text":"hello"}`,
		},
		{
			msg:        "ok (with query)",
			path:       "/articles/1/comments?loud=true",
			code:       200,
			body:       strings.NewReader(`{"text": "hello"}`),
			resultBody: `{"author":"someone","text":"HELLO"}`,
		},
		{
			msg:        "ok (invalid query is ignored)",
			path:       "/articles/1/comments?loud=ababba",
			code:       200,
			body:       strings.NewReader(`{"text": "hello"}`),
			resultBody: `{"author":"someone","text":"hello"}`,
		},
		{
			msg:        "ok (unsupported query is ignored)",
			path:       "/articles/1/comments?unsupported=true",
			code:       200,
			body:       strings.NewReader(`{"text": "hello"}`),
			resultBody: `{"author":"someone","text":"hello"}`,
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

			if want := strings.TrimSpace(c.resultBody); want != "" {
				var ob interface{}
				if err := json.NewDecoder(res.Body).Decode(&ob); err != nil {
					t.Errorf("unexpected decode error: %+v", err)
				}
				var buf strings.Builder
				if err := json.NewEncoder(&buf).Encode(ob); err != nil {
					t.Errorf("unexpected encode error: %+v", err)
				}
				got := strings.TrimSpace(buf.String())
				if want != got {
					t.Errorf("want response body is %v but got %v", want, got)
				}
			}
		})
	}
}
