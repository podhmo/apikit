package emitfile

import (
	"path/filepath"
	"testing"
)

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
		{ // with modify
			rootdir: "/go/src/my-project/with-child",
			pkgpath: "/foo/bar",
			want:    filepath.Join("/go/foo", "bar"),
			modify: func(r *PathResolver) {
				r.Add("/foo/bar", "/go/foo/bar")
			},
		},
		{
			rootdir: "/go/src/my-project/with-parent",
			pkgpath: "/foo/bar",
			want:    filepath.Join("/go/foo", "bar"),
			modify: func(r *PathResolver) {
				r.Add("/foo", "/go/foo")
			},
		},
		{
			rootdir: "/go/src/my-project/with-parent-child/parent",
			pkgpath: "/foo/bee/boo",
			want:    filepath.Join("/go/foo", "bee/boo"),
			modify: func(r *PathResolver) {
				r.Add("/foo", "/go/foo") // use-this
				r.Add("/foo/bar", "/go/bar")
			},
		},
		{
			rootdir: "/go/src/my-project/with-parent-child/child",
			pkgpath: "/foo/bar/boo",
			want:    filepath.Join("/go/bar", "boo"),
			modify: func(r *PathResolver) {
				r.Add("/foo", "/go/foo")
				r.Add("/foo/bar", "/go/bar") // use-this
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.pkgpath, func(t *testing.T) {
			r := newPathResolver(c.rootdir)
			if c.modify != nil {
				c.modify(r)
			}
			got, err := r.Resolve(c.pkgpath)
			if err != nil {
				t.Fatalf("unexpected error %+v", err)
			}
			if want := c.want; want != got {
				t.Errorf("want:\n\t%q\nbut got (with rootdir=%q):\n\t%q", want, c.rootdir, got)
			}
		})
	}
}
