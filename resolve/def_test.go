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
			Def:  resolver.Def(ListUser),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params primitive pointer -> primitive-pointer
			Def:  resolver.Def(ListUserWithPrimitivePointer),
			Args: []Kind{KindComponent, KindPrimitivePointer, KindPrimitivePointer},
		},
		{
			// OK: handle params context.Context -> ignored
			Def:  resolver.Def(ListUserWithContext),
			Args: []Kind{KindIgnored, KindComponent},
		},
		{
			// OK: handle params function -> component
			Def:  resolver.Def(ListUserWithFunction),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params interface -> component
			Def:  resolver.Def(ListUserWithInterface),
			Args: []Kind{KindComponent},
		},
		{
			// OK: handle params struct -> data
			Def:  resolver.Def(GetUserWithStruct),
			Args: []Kind{KindComponent, KindData},
		},
		{
			// OK: handle params primitive -> primitive
			Def:  resolver.Def(GetUserWithPrimitive),
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
