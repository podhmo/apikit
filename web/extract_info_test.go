package web_test

import (
	"reflect"
	"testing"

	"github.com/podhmo/apikit/web"
	reflectshape "github.com/podhmo/reflect-shape"
	"github.com/podhmo/reflect-shape/arglist"
)

type Bar struct{}

func getFooBar(fooId string, barId int) (*Bar, error) {
	return nil, nil
}

func TestExtractPathInfo(t *testing.T) {
	extractor := reflectshape.NewExtractor()
	extractor.ArglistLookup = arglist.NewLookup()

	cases := []struct {
		msg           string
		shape         reflectshape.Shape
		variableNames []string

		wantName     string
		wantArgTypes []string
		//wantErr error
	}{
		{
			msg:           "ok",
			shape:         extractor.Extract(getFooBar),
			variableNames: []string{"fooId", "barId"},
			wantName:      "getFooBar",
			wantArgTypes:  []string{"string", "int"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			got, err := web.ExtractPathInfo(c.variableNames, c.shape)
			if err != nil {
				t.Fatalf("unexpected error %+v", err)
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
