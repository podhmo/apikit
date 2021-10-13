package tinypkg

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBindingSort(t *testing.T) {
	pkg := NewPackage("pkg", "")
	mustBinding := func(name string, f *Func) *Binding {
		b, err := NewBinding(name, f)
		if err != nil {
			panic(err)
		}
		return b
	}
	cases := []struct {
		msg    string
		input  BindingList
		output []string
	}{
		{
			msg:    "empty",
			input:  nil,
			output: nil,
		},
		{
			msg:    "one",
			input:  BindingList{mustBinding("X", pkg.NewFunc("getX", nil, []*Var{{Node: pkg.NewSymbol("X")}}))},
			output: []string{"X <- getX(...)"},
		},
		{
			msg: "two,no-deps",
			input: BindingList{
				mustBinding("X", pkg.NewFunc("getX", nil, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
			},
			output: []string{
				"X <- getX(...)",
				"Y <- getY(...)",
			},
		},
		// {
		// 	msg: "ng-two,deps",
		// 	input: BindingList{
		// 		mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "MISSING", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
		// 		mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
		// 	},
		// 	output: []string{
		// 		"Y <- getY(...)",
		// 		"X <- getX(...)",
		// 	},
		// },
		{
			msg: "two,deps",
			input: BindingList{
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
			},
			output: []string{
				"Y <- getY(...)",
				"X <- getX(...)",
			},
		},
		{
			msg: "two,deps,reorder",
			input: BindingList{
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
			},
			output: []string{
				"Y <- getY(...)",
				"X <- getX(...)",
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			sorted, err := c.input.TopologicalSorted()
			if err != nil {
				t.Fatalf("unexpected error %+v", err)
			}

			var got []string
			for _, b := range sorted {
				got = append(got, fmt.Sprintf("%s <- %s(...)", b.Name, b.Provider.Name))
			}
			if want := c.output; !reflect.DeepEqual(want, got) {
				t.Errorf("want:\n\t%v\nbut got:\n\t%v", want, got)
			}
		})
	}
}
