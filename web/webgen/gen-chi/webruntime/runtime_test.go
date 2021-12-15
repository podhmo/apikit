package webruntime_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"webruntime"

	"github.com/go-chi/chi/v5"
)

func TestPathParams(t *testing.T) {
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
