package webruntime

import (
	"reflect"
	"testing"
)

func TestScroll(t *testing.T) {
	type Row struct {
		ID    int
		Value int
	}
	rows := []Row{
		{ID: 1, Value: 10}, {ID: 2, Value: 20}, {ID: 3, Value: 30},
		{ID: 4, Value: 40}, {ID: 5, Value: 50}, {ID: 6, Value: 60},
		{ID: 7, Value: 70}, {ID: 8, Value: 80}, {ID: 9, Value: 90},
		{ID: 10, Value: 100},
	}
	sc := &ScrollContext{
		More:     true,
		Size:     3,
		LatestID: nil,
		Key:      "ID",
	}

	fetch := func(sc *ScrollContext) []Row {
		if !sc.More {
			return nil
		}
		i := 0
		size := sc.Size + 1
		latestID := sc.LatestID
		r := make([]Row, 0, size)
		for _, row := range rows {
			if latestID == nil || *latestID < row.ID {
				r = append(r, row)
				i++
			}
			if i >= size {
				break
			}
		}
		return r
	}

	var scrollID string
	type want struct {
		result []Row
		more   bool
	}
	t.Run("page=0", func(t *testing.T) {
		want := want{
			result: []Row{
				{ID: 1, Value: 10}, {ID: 2, Value: 20}, {ID: 3, Value: 30},
			},
			more: true,
		}

		result := fetch(sc)
		ns, err := sc.NextState(result)
		if err != nil {
			t.Fatalf("unexpected error, bindState: %+v", err)
		}
		if ns.More {
			result = result[:len(result)-1]
		}

		if want, got := want.result, result; !reflect.DeepEqual(want, got) {
			t.Errorf("result, want is %v, but got is %v", want, got)
		}
		if want, got := want.more, ns.More; want != got {
			t.Errorf(".More, want is %v, but got is %v", want, got)
		}

		if ns.ScrollID == "" {
			t.Error("scrollId must not be empty")
		}
		scrollID = ns.ScrollID
	})
	t.Run("page=1", func(t *testing.T) {
		want := want{
			result: []Row{
				{ID: 4, Value: 40}, {ID: 5, Value: 50}, {ID: 6, Value: 60},
			},
			more: true,
		}

		var sc ScrollContext
		if err := sc.Decode(scrollID); err != nil {
			t.Errorf("unexpected error, decode context: %+v", err)
		}

		result := fetch(&sc)
		ns, err := sc.NextState(result)
		if err != nil {
			t.Fatalf("unexpected error, bindState: %+v", err)
		}
		if ns.More {
			result = result[:len(result)-1]
		}

		if want, got := want.result, result; !reflect.DeepEqual(want, got) {
			t.Errorf("result, want is %v, but got is %v", want, got)
		}
		if want, got := want.more, ns.More; want != got {
			t.Errorf(".More, want is %v, but got is %v", want, got)
		}

		if ns.ScrollID == "" {
			t.Error("scrollId must not be empty")
		}
		scrollID = ns.ScrollID
	})
	t.Run("page=3", func(t *testing.T) {
		want := want{
			result: []Row{
				{ID: 7, Value: 70}, {ID: 8, Value: 80}, {ID: 9, Value: 90},
			},
			more: true,
		}

		var sc ScrollContext
		if err := sc.Decode(scrollID); err != nil {
			t.Errorf("unexpected error, decode context: %+v", err)
		}

		result := fetch(&sc)
		ns, err := sc.NextState(result)
		if err != nil {
			t.Fatalf("unexpected error, bindState: %+v", err)
		}
		if ns.More {
			result = result[:len(result)-1]
		}

		if want, got := want.result, result; !reflect.DeepEqual(want, got) {
			t.Errorf("result, want is %v, but got is %v", want, got)
		}
		if want, got := want.more, ns.More; want != got {
			t.Errorf(".More, want is %v, but got is %v", want, got)
		}

		if ns.ScrollID == "" {
			t.Error("scrollId must not be empty")
		}
		scrollID = ns.ScrollID
	})
	t.Run("page=4", func(t *testing.T) {
		want := want{
			result: []Row{
				{ID: 10, Value: 100},
			},
			more: false,
		}

		var sc ScrollContext
		if err := sc.Decode(scrollID); err != nil {
			t.Errorf("unexpected error, decode context: %+v", err)
		}

		result := fetch(&sc)
		ns, err := sc.NextState(result)
		if err != nil {
			t.Fatalf("unexpected error, bindState: %+v", err)
		}
		if ns.More {
			result = result[:len(result)-1]
		}

		if want, got := want.result, result; !reflect.DeepEqual(want, got) {
			t.Errorf("result, want is %v, but got is %v", want, got)
		}
		if want, got := want.more, ns.More; want != got {
			t.Errorf(".More, want is %v, but got is %v", want, got)
		}

		if ns.ScrollID == "" {
			t.Error("scrollId must not be empty")
		}
		scrollID = ns.ScrollID
	})
	t.Run("page=5", func(t *testing.T) {
		want := want{
			result: nil,
			more:   false,
		}

		var sc ScrollContext
		if err := sc.Decode(scrollID); err != nil {
			t.Errorf("unexpected error, decode context: %+v", err)
		}

		result := fetch(&sc)
		ns, err := sc.NextState(result)
		if err != nil {
			t.Fatalf("unexpected error, bindState: %+v", err)
		}
		if ns.More {
			result = result[:len(result)-1]
		}

		if want, got := want.result, result; !reflect.DeepEqual(want, got) {
			t.Errorf("result, want is %v, but got is %v", want, got)
		}
		if want, got := want.more, ns.More; want != got {
			t.Errorf(".More, want is %v, but got is %v", want, got)
		}

		if ns.ScrollID == "" {
			t.Error("scrollId must not be empty")
		}
		scrollID = ns.ScrollID
	})
}
