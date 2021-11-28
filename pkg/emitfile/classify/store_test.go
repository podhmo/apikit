package classify

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/podhmo/apikit/pkg/difftest"
)

func TestStore(t *testing.T) {
	newEntry := func(name string) *FileEntry {
		return &FileEntry{fileinfo: fileinfo{Name: name, Hash: []byte(name)}}
	}
	created := func(name string) Result {
		return Result{Type: ResultTypeCreate, Entry: newEntry(name)}
	}
	updated := func(name string) Result {
		return Result{Type: ResultTypeUpdate, Entry: newEntry(name)}
	}
	deleted := func(name string) Result {
		return Result{Type: ResultTypeDelete, Entry: newEntry(name)}
	}
	notchanged := func(name string) Result {
		return Result{Type: ResultTypeNotChanged, Entry: newEntry(name)}
	}
	mustTime := func(s string) time.Time {
		mtime, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatalf("!! %+v", err)
		}
		return mtime
	}

	// TODO: first time load

	stages := []struct {
		results []Result
		mtime   time.Time
		want    string
	}{
		{
			results: []Result{created("hello.txt"), created("byebye.txt")},
			mtime:   mustTime("2000-01-01T00:00:00Z"),
			want: `
[
	{
		"name": "hello.txt",
		"hash": "aGVsbG8udHh0",
		"mtime": "2000-01-01T00:00:00Z"
	},
	{
		"name": "byebye.txt",
		"hash": "YnllYnllLnR4dA==",
		"mtime": "2000-01-01T00:00:00Z"
	}
]`,
		},
		{
			results: []Result{updated("hello.txt"), deleted("byebye.txt"), created("yoo.txt")},
			mtime:   mustTime("2000-02-01T00:00:00Z"),
			want: `
[
	{
		"name": "hello.txt",
		"hash": "aGVsbG8udHh0",
		"mtime": "2000-02-01T00:00:00Z"
	},
	{
		"name": "yoo.txt",
		"hash": "eW9vLnR4dA==",
		"mtime": "2000-02-01T00:00:00Z"
	}
]`,
		},
		{
			results: []Result{updated("hello.txt"), notchanged("yoo.txt")},
			mtime:   mustTime("2000-03-01T00:00:00Z"),
			want: `
[
	{
		"name": "hello.txt",
		"hash": "aGVsbG8udHh0",
		"mtime": "2000-03-01T00:00:00Z"
	},
	{
		"name": "yoo.txt",
		"hash": "eW9vLnR4dA==",
		"mtime": "2000-02-01T00:00:00Z"
	}
]`,
		},
	}

	t.Run("save0", func(t *testing.T) {
		i := 0
		s := &JSONFileStore{Mtime: stages[i].mtime}
		results := stages[i].results

		var buf bytes.Buffer
		if err := s.WriteData(&buf, results); err != nil {
			t.Errorf("unexpected write data: %+v", err)
		}

		want := stages[i].want
		if want, got := strings.TrimSpace(want), strings.TrimSpace(buf.String()); want != got {
			difftest.LogDiffGotStringAndWantString(t, got, want)
		}
	})

	t.Run("load1", func(t *testing.T) {
		i := 1
		s := &JSONFileStore{Mtime: stages[i].mtime}
		entries, err := s.ReadData(bytes.NewBufferString(stages[i-1].want))
		if err != nil {
			t.Errorf("unexpected read data: %+v", err)
		}

		want := make([]string, 0, len(stages[i-1].results))
		for _, x := range stages[i-1].results {
			if x.Type == ResultTypeDelete {
				continue
			}
			want = append(want, x.Name())
		}

		got := make([]string, 0, len(entries))
		for _, x := range entries {
			got = append(got, x.Name())
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("want:\n\t%v\nbut got:\n\t%v", want, got)
		}
	})

	t.Run("save1", func(t *testing.T) {
		i := 1
		s := &JSONFileStore{Mtime: stages[i].mtime}

		results := make([]Result, len(stages[i].results))
		for j, x := range stages[i].results {
			entry := x.Entry.(*FileEntry)
			entry.fileinfo.Mtime = stages[i-1].mtime
			results[j] = x
		}

		var buf bytes.Buffer
		if err := s.WriteData(&buf, results); err != nil {
			t.Errorf("unexpected write data: %+v", err)
		}

		want := stages[i].want
		if want, got := strings.TrimSpace(want), strings.TrimSpace(buf.String()); want != got {
			difftest.LogDiffGotStringAndWantString(t, got, want)
		}
	})

	t.Run("load2", func(t *testing.T) {
		i := 2
		s := &JSONFileStore{Mtime: stages[i].mtime}
		entries, err := s.ReadData(bytes.NewBufferString(stages[i-1].want))
		if err != nil {
			t.Errorf("unexpected read data: %+v", err)
		}

		want := make([]string, 0, len(stages[i-1].results))
		for _, x := range stages[i-1].results {
			if x.Type == ResultTypeDelete {
				continue
			}
			want = append(want, x.Name())
		}

		got := make([]string, 0, len(entries))
		for _, x := range entries {
			got = append(got, x.Name())
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("want:\n\t%v\nbut got:\n\t%v", want, got)
		}
	})

	t.Run("save2", func(t *testing.T) {
		i := 2
		s := &JSONFileStore{Mtime: stages[i].mtime}

		results := make([]Result, len(stages[i].results))
		for j, x := range stages[i].results {
			entry := x.Entry.(*FileEntry)
			entry.fileinfo.Mtime = stages[i-1].mtime
			results[j] = x
		}

		var buf bytes.Buffer
		if err := s.WriteData(&buf, results); err != nil {
			t.Errorf("unexpected write data: %+v", err)
		}

		want := stages[i].want
		if want, got := strings.TrimSpace(want), strings.TrimSpace(buf.String()); want != got {
			difftest.LogDiffGotStringAndWantString(t, got, want)
		}
	})
}
