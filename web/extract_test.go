package web_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/podhmo/apikit/resolve"
	"github.com/podhmo/apikit/web"
)

type Bar struct{}

func getFooBar(ctx context.Context, fooId string, barId int) (*Bar, error) {
	return nil, nil
}

func TestExtractPathInfo(t *testing.T) {
	resolver := resolve.NewResolver()
	cases := []struct {
		msg           string
		fn            interface{}
		variableNames []string

		wantName     string
		wantArgTypes []string
		wantErr      error
	}{
		{
			msg:           "ok",
			fn:            getFooBar,
			variableNames: []string{"fooId", "barId"},
			wantName:      "getFooBar",
			wantArgTypes:  []string{"string", "int"},
		},
		{
			msg:           "num-of-varnames-is-smaller",
			fn:            getFooBar,
			variableNames: []string{"fooId"},
			wantErr:       web.ErrMismatchNumberOfVariables,
		},
		{
			msg:           "num-of-varnames-is-bigger",
			fn:            getFooBar,
			variableNames: []string{"fooId", "barId", "bazId"},
			wantErr:       web.ErrMismatchNumberOfVariables,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			def := resolver.Def(c.fn)
			got, err := web.ExtractPathInfo(c.variableNames, def)
			if c.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error %+v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("must be error is occured (want error is %s)", c.wantErr)
				}
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("unexpected error %+v (want error is %s", err, c.wantErr)
				}
				return
			}

			if want, got := c.wantName, got.Name; want != got {
				t.Errorf("want name\n\t%q\nbut got\n%q", want, got)
			}

			var gotArgTypes []string
			for _, v := range got.Variables {
				gotArgTypes = append(gotArgTypes, v.Shape.GetName())
			}
			if want, got := c.wantArgTypes, gotArgTypes; !reflect.DeepEqual(want, got) {
				t.Errorf("want arg types\n\t%v\nbut got\n%v", want, got)
			}
		})
	}
}
