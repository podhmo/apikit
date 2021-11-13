package classify

import (
	"bytes"
	"fmt"
)

type Entry interface {
	Name() string
	Hash() ([]byte, error)
}

type funcEntry struct {
	name     string
	hashFunc func() ([]byte, error)
}

func NewEntry(name string, hashFunc func() ([]byte, error)) Entry {
	return &funcEntry{
		name:     name,
		hashFunc: hashFunc,
	}
}
func (e *funcEntry) Name() string {
	return e.name
}
func (e *funcEntry) Hash() ([]byte, error) {
	return e.hashFunc()
}

// TODO: concurrent

func Classify(
	prevEntries, entries []Entry,
) ([]Result, error) {
	Results := make([]Result, len(entries))

	prevUsedCounter := make([]int, len(prevEntries))
	hasPrev := len(prevEntries) > 0

	for i, entry := range entries {
		if !hasPrev {
			Results[i] = Result{Type: ResultTypeCreate, Entry: entry}
			continue
		}

		var prev Entry
		name := entry.Name()
		for j, x := range prevEntries {
			if x.Name() == name {
				prevUsedCounter[j]++
				prev = x
				break
			}
		}

		if prev == nil {
			Results[i] = Result{Type: ResultTypeCreate, Entry: entry}
			continue
		}

		currentHash, err := entry.Hash()
		if err != nil {
			return nil, fmt.Errorf("get hash of current entry=%s: %w", name, err)
		}
		prevHash, err := prev.Hash()
		if err != nil {
			return nil, fmt.Errorf("get hash of previous entry=%s: %w", name, err)
		}
		if bytes.Equal(currentHash, prevHash) {
			Results[i] = Result{Type: ResultTypeNotChanged, Entry: entry}
			continue
		}
		Results[i] = Result{Type: ResultTypeUpdate, Entry: entry}
	}

	if hasPrev {
		for i, c := range prevUsedCounter {
			if c == 0 {
				entry := prevEntries[i]
				Results = append(Results, Result{Type: ResultTypeDelete, Entry: entry})
			}
		}
	}
	return Results, nil
}

type ResultType string

const (
	ResultTypeUNKNOWN     ResultType = ""
	ResultTypeCreate      ResultType = "C"
	ResultTypeUpdate      ResultType = "U"
	ResultTypeDelete      ResultType = "D"
	ResultTypeNotChanged ResultType = "-"
)

type Result struct {
	Type  ResultType
	Entry Entry
}

func (a Result) Name() string {
	return a.Entry.Name()
}

func (a Result) String() string {
	return fmt.Sprintf("%s %s", a.Type, a.Entry.Name())
}
