package emitfile

import (
	"path/filepath"
	"testing"
)

type fakeLogger struct{}

func (l *fakeLogger) Printf(fmt string, args ...interface{}) {}

func TestPathResolver(t *testing.T) {
	cases := []struct {
		rootdir string
		pkgpath string
		want    string

		hasError bool
		modify   func(*PathResolver)
	}{
		{ // relative
			rootdir: "/go/src/my-project",
			pkgpath: "foo",
			want:    filepath.Join("/go/src/my-project", "foo"),
		},
		{
			rootdir: "/go/src/my-project",
			pkgpath: "foo/bar",
			want:    filepath.Join("/go/src/my-project", "foo/bar"),
		},
		{
			rootdir: "/go/src/my-project",
			pkgpath: "../bar",
			want:    filepath.Join("/go/src", "bar"),
		},
		{
			rootdir:  "/go/src/my-project",
			pkgpath:  "xxx/yyy/not-found",
			hasError: false,
			want:     filepath.Join("/go/src/my-project", "xxx/yyy/not-found"), // always use rootdir
		},
		{ // absolute
			rootdir: "/go/src/my-project",
			pkgpath: "/",
			want:    filepath.Join("/go/src/my-project", ""),
		},
		{
			rootdir: "/go/src/my-project",
			pkgpath: "/foo/bar",
			want:    filepath.Join("/go/src/my-project", "foo/bar"),
		},
		{
			rootdir:  "/go/src/my-project",
			pkgpath:  "/xxx/yyy/not-found",
			hasError: false,
			want:     filepath.Join("/go/src/my-project", "xxx/yyy/not-found"), // always use rootdir
		},
		{ // with modify
			rootdir: "/go/src/my-project/with-child",
			pkgpath: "/foo/bar",
			want:    filepath.Join("/go/foo", "bar"),
			modify: func(r *PathResolver) {
				r.AddRoot("/foo/bar", "/go/foo/bar")
			},
		},
		{
			rootdir: "/go/src/my-project/with-parent",
			pkgpath: "/foo/bar",
			want:    filepath.Join("/go/foo", "bar"),
			modify: func(r *PathResolver) {
				r.AddRoot("/foo", "/go/foo")
			},
		},
		{
			rootdir: "/go/src/my-project/with-parent-child/parent",
			pkgpath: "/foo/bee/boo",
			want:    filepath.Join("/go/foo", "bee/boo"),
			modify: func(r *PathResolver) {
				r.AddRoot("/foo", "/go/foo") // use-this
				r.AddRoot("/foo/bar", "/go/bar")
			},
		},
		{
			rootdir: "/go/src/my-project/with-parent-child/child",
			pkgpath: "/foo/bar/boo",
			want:    filepath.Join("/go/bar", "boo"),
			modify: func(r *PathResolver) {
				r.AddRoot("/foo", "/go/foo")
				r.AddRoot("/foo/bar", "/go/bar") // use-this
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.pkgpath, func(t *testing.T) {
			r := newPathResolver(c.rootdir, &Config{Log: &fakeLogger{}})
			if c.modify != nil {
				c.modify(r)
			}
			got, err := r.ResolvePath(c.pkgpath)
			if c.hasError && err == nil {
				t.Fatalf("need error, but return nil")
			} else if err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			if want := c.want; want != got {
				t.Errorf("want:\n\t%q\nbut got (with rootdir=%q):\n\t%q", want, c.rootdir, got)
			}
		})
	}
}
