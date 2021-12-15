package webruntime_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"webruntime"

	"github.com/go-chi/chi/v5"
)

func TestBindPathParams(t *testing.T) {
	type Data struct {
		FooID string `path:"fooId"`
		BarID int    `path:"barId"`

		err error
	}

	want := Data{FooID: "1", BarID: 10}

	var data *Data
	r := chi.NewRouter()
	r.Get("/foo/{fooId}/bar/{barId}", func(w http.ResponseWriter, req *http.Request) {
		if err := webruntime.BindPathParams(data, req, "fooId", "barId"); err != nil {
			data.err = err
		}
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	t.Run("ok", func(t *testing.T) {
		path := "/foo/1/bar/10"
		data = &Data{}
		if _, err := http.Get(ts.URL + path); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}

		if want, got := want.FooID, data.FooID; want != got {
			t.Errorf("want foo id is %q, but got is %q", want, got)
		}
		if want, got := want.BarID, data.BarID; want != got {
			t.Errorf("want bar id is %v, but got is %v", want, got)
		}
	})

	t.Run("ng", func(t *testing.T) {
		path := "/foo/foo/bar/bar"
		data = &Data{}
		if _, err := http.Get(ts.URL + path); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}

		if want, got := want.FooID, data.FooID; !(want != got) {
			t.Errorf("want foo id is %q, but got is %q", want, got)
		}
		if want, got := want.BarID, data.BarID; !(want != got) {
			t.Errorf("want bar id is %v, but got is %v", want, got)
		}
	})
}

func TestBindQuery(t *testing.T) {
	type Data struct {
		Verbose *bool `query:"verbose"`
		Limit   *int  `limit:"limit"`
	}

	t.Run("ok", func(t *testing.T) {
		var data Data
		req := httptest.NewRequest("GET", "/?verbose=true&limit=10&xxx=yyyy", nil)
		webruntime.BindQuery(&data, req)

		verbose := true
		limit := 10
		want := Data{Verbose: &verbose, Limit: &limit}
		if want, got := want.Verbose, data.Verbose; *want != *got {
			t.Errorf("want verbose is %v, but got is %v", *want, *got)
		}
		if want, got := want.Limit, data.Limit; *want != *got {
			t.Errorf("want limit is %v, but got is %v", *want, *got)
		}
	})
	t.Run("invalid ignored", func(t *testing.T) {
		var data Data
		req := httptest.NewRequest("GET", "/?verbose=abbaba&limit=fooo", nil)
		webruntime.BindQuery(&data, req)

		verbose := false
		limit := 0
		want := Data{Verbose: &verbose, Limit: &limit}
		if want, got := want.Verbose, data.Verbose; *want != *got {
			t.Errorf("want verbose is %v, but got is %v", *want, *got)
		}
		if want, got := want.Limit, data.Limit; *want != *got {
			t.Errorf("want limit is %v, but got is %v", *want, *got)
		}
	})
}

func TestBindHeader(t *testing.T) {
	type Data struct {
		APIKey  *string `header:"API_KEY"`
		Verbose *bool   `header:"VERBOSE"`
	}

	t.Run("ok", func(t *testing.T) {
		var data Data
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "xxx")
		req.Header.Set("VERBOSE", "1")
		req.Header.Set("XXXX", "YYY")

		webruntime.BindHeader(&data, req)

		apiKey := "xxx"
		verbose := true
		want := Data{Verbose: &verbose, APIKey: &apiKey}
		if want, got := want.APIKey, data.APIKey; *want != *got {
			t.Errorf("want apikey is %v, but got is %v", *want, *got)
		}
		if want, got := want.Verbose, data.Verbose; *want != *got {
			t.Errorf("want verbose is %v, but got is %v", *want, *got)
		}
	})

	t.Run("invalid ignored", func(t *testing.T) {
		var data Data
		req := httptest.NewRequest("GET", "/", nil)

		webruntime.BindHeader(&data, req)
		if got := data.APIKey; got != nil {
			t.Errorf("want apikey is nil, but got is %q", *got)
		}
		if got := data.Verbose; got != nil {
			t.Errorf("want verbose is nil, but got is %v", *got)
		}
	})
}

func TestBindBody(t *testing.T) {
	type Data struct {
		APIKey  string `json:"apikey"`
		Verbose bool   `json:"verbose"`
	}

	t.Run("ok", func(t *testing.T) {
		var data Data

		body := `{"apikey": "xxx", "verbose": true, "xxx": "yyy"}`
		if err := webruntime.BindBody(&data, ioutil.NopCloser(strings.NewReader(body))); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}

		want := Data{Verbose: true, APIKey: "xxx"}
		if want, got := want.APIKey, data.APIKey; want != got {
			t.Errorf("want apikey is %v, but got is %v", want, got)
		}
		if want, got := want.Verbose, data.Verbose; want != got {
			t.Errorf("want verboseis %v, but got is %v", want, got)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var data Data

		body := `{"verbose": "foo"`
		if err := webruntime.BindBody(&data, ioutil.NopCloser(strings.NewReader(body))); err == nil {
			t.Errorf("error is expected but nil")
		}

		want := Data{}
		if want, got := want.APIKey, data.APIKey; want != got {
			t.Errorf("want apikey is %v, but got is %v", want, got)
		}
		if want, got := want.Verbose, data.Verbose; want != got {
			t.Errorf("want verboseis %v, but got is %v", want, got)
		}
	})

	t.Run("invalid field type", func(t *testing.T) {
		var data Data

		body := `{"verbose": "foo"}`
		if err := webruntime.BindBody(&data, ioutil.NopCloser(strings.NewReader(body))); err == nil {
			t.Errorf("error is expected but nil")
		}

		want := Data{}
		if want, got := want.APIKey, data.APIKey; want != got {
			t.Errorf("want apikey is %v, but got is %v", want, got)
		}
		if want, got := want.Verbose, data.Verbose; want != got {
			t.Errorf("want verboseis %v, but got is %v", want, got)
		}
	})
}
