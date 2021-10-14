package tinypkg

import (
	"fmt"
	"reflect"
	"strings"
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
		msg   string
		input BindingList

		output []string
		hasErr bool
	}{
		{
			msg:    "empty",
			input:  nil,
			output: nil,
		},
		{
			msg:    "one",
			input:  BindingList{mustBinding("X", pkg.NewFunc("getX", nil, []*Var{{Node: pkg.NewSymbol("X")}}))},
			output: []string{"X <- getX()"},
		},
		{
			msg: "two,no-deps",
			input: BindingList{
				mustBinding("X", pkg.NewFunc("getX", nil, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
			},
			output: []string{
				"X <- getX()",
				"Y <- getY()",
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
				"Y <- getY()",
				"X <- getX(Y)",
			},
		},
		{
			msg: "two,deps,reorder",
			input: BindingList{
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
			},
			output: []string{
				"Y <- getY()",
				"X <- getX(Y)",
			},
		},
		{
			msg: "three,deps,shared",
			input: BindingList{
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Z", pkg.NewFunc("getZ", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("Z")}})),
			},
			output: []string{
				"Y <- getY()",
				"X <- getX(Y)",
				"Z <- getZ(Y)",
			},
		},
		{
			msg: "three,deps,chained",
			input: BindingList{
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Z", pkg.NewFunc("getZ", []*Var{{Name: "X", Node: pkg.NewSymbol("X")}}, []*Var{{Node: pkg.NewSymbol("Z")}})),
			},
			output: []string{
				"Y <- getY()",
				"X <- getX(Y)",
				"Z <- getZ(X)",
			},
		},
		{
			msg: "three,deps,chained,reorder",
			input: BindingList{
				mustBinding("Z", pkg.NewFunc("getZ", []*Var{{Name: "X", Node: pkg.NewSymbol("X")}}, []*Var{{Node: pkg.NewSymbol("Z")}})),
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Y", pkg.NewFunc("getY", nil, []*Var{{Node: pkg.NewSymbol("Y")}})),
			},
			output: []string{
				"Y <- getY()",
				"X <- getX(Y)",
				"Z <- getZ(X)",
			},
		},
		{
			msg: "ng-three,deps,circular-dependency",
			input: BindingList{
				mustBinding("Z", pkg.NewFunc("getZ", []*Var{{Name: "X", Node: pkg.NewSymbol("X")}}, []*Var{{Node: pkg.NewSymbol("Z")}})),
				mustBinding("X", pkg.NewFunc("getX", []*Var{{Name: "Y", Node: pkg.NewSymbol("Y")}}, []*Var{{Node: pkg.NewSymbol("X")}})),
				mustBinding("Y", pkg.NewFunc("getY", []*Var{{Name: "Z", Node: pkg.NewSymbol("Z")}}, []*Var{{Node: pkg.NewSymbol("Y")}})),
			},
			hasErr: true,
			output: []string{
				"Y <- getY(Z)",
				"X <- getX(Y)",
				"Z <- getZ(X)",
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.msg, func(t *testing.T) {
			sorted, err := c.input.TopologicalSorted()
			if c.hasErr {
				if err == nil {
					t.Error("expected error, but not occured")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}

			var got []string
			for _, b := range sorted {
				args := b.argsAliases
				if args == nil {
					for _, x := range b.Provider.Args {
						args = append(args, x.Name)
					}
				}
				got = append(got, fmt.Sprintf("%s <- %s(%s)", b.Name, b.Provider.Name, strings.Join(args, ", ")))
			}
			if want := c.output; !reflect.DeepEqual(want, got) {
				t.Errorf("want:\n\t%v\nbut got:\n\t%v", want, got)
			}
		})
	}
}
