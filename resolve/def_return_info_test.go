package resolve

import (
	"errors"
	"testing"
)

func TestBindReturnInfo(t *testing.T) {
	resolver := NewResolver()
	cases := []struct {
		msg string
		fn  interface{}

		hasError   bool
		hasCleanup bool

		wantErr error
	}{
		{
			msg: "ok-return-0",
			fn:  func() {},
		},
		{
			msg: "ok-return-1",
			fn:  func() interface{} { return nil },
		},
		{
			msg:      "ok-return-1--error",
			fn:       func() error { return nil },
			hasError: true,
		},
		{
			msg:        "ok-return-1--cleanup",
			fn:         func() func() { return nil },
			hasCleanup: true,
		},
		{
			msg:      "ok-return-2--error",
			fn:       func() (interface{}, error) { return nil, nil },
			hasError: true,
		},
		{
			msg:        "ok-return-2--cleanup",
			fn:         func() (interface{}, func()) { return nil, nil },
			hasCleanup: true,
		},
		{
			msg:     "ng-return-2",
			fn:      func() (interface{}, interface{}) { return nil, nil },
			wantErr: ErrUnexpectedReturnType,
		},
		{
			msg:        "ok-return-3",
			fn:         func() (interface{}, func(), error) { return nil, nil, nil },
			hasError:   true,
			hasCleanup: true,
		},
		{
			msg:     "ng-return-3",
			fn:      func() (interface{}, error, func()) { return nil, nil, nil },
			wantErr: ErrUnexpectedReturnType,
		},
		{
			msg:     "ng-return-4",
			fn:      func() (interface{}, interface{}, func(), error) { return nil, nil, nil, nil },
			wantErr: ErrUnexpectedReturnType,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			def := resolver.Def(c.fn)
			err := bindReturnsInfo(def)
			if c.wantErr != nil {
				if err == nil {
					t.Fatalf("must be error is occured (want error is %+v)", c.wantErr)
				}
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("unexpected error %+v (want error is %+v)", err, c.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error %+v", err)
				}
			}

			if want, got := c.hasError, def.HasError; want != got {
				t.Errorf("want hasError:%v, but got:%v", want, got)
			}
			if want, got := c.hasCleanup, def.HasCleanup; want != got {
				t.Errorf("want hasCleanup:%v, but got:%v", want, got)
			}
		})
	}
}
