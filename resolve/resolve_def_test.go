package resolve

import (
	"testing"
)

// TODO: handle returns

func Test(t *testing.T) {
	resolver := NewResolver()
	cases := []struct {
		Def  *Def
		Args []Kind
	}{
		{
			// OK: handle params pointer -> component
			Def:  resolver.Resolve(ListUser),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params context.Context -> ignored
			Def:  resolver.Resolve(ListUserWithContext),
			Args: []Kind{KindIgnored, KindComponent},
		},
		{
			// OK: handle params function -> component
			Def:  resolver.Resolve(ListUserWithFunction),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params interface -> component
			Def:  resolver.Resolve(ListUserWithInterface),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params struct -> data
			Def:  resolver.Resolve(GetUserWithStruct),
			Args: []Kind{KindComponent, KindData},
		},
		{
			// OK: handle params primitive -> primitive
			Def:  resolver.Resolve(GetUserWithPrimitive),
			Args: []Kind{KindComponent, KindPrimitive},
		},
	}

	for _, c := range cases {
		c := c
		Def := c.Def

		t.Run(Def.Name, func(t *testing.T) {
			t.Logf("input    : %+v", Def.GoString())
			t.Logf("signature: %+v", Def.Shape.GetReflectType())
			t.Logf("want args: %+v", c.Args)
			for i, x := range Def.Args {
				wantKind := c.Args[i]
				gotKind := x.Kind
				t.Run(x.Name, func(t *testing.T) {
					if wantKind != gotKind {
						t.Errorf("want %s but got %s", wantKind, gotKind)
					}
				})
			}
		})
	}
}
