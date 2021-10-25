// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runtime

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
)

type ScrollContext struct {
	Key     string   `json:"key"`     // e.g. pk
	OrderBy []string `json:"orderBy"` // e.g. {-createdAt, -pk}

	LatestID *ScrollT `json:"latestId"`
	More     bool     `json:"more"`
	Size     int      `json:"size"`

	Additional interface{} `json:"additional"` // additionalContext
}

func NewScrollContext(key string, size int) *ScrollContext {
	return &ScrollContext{
		Key:  key,
		More: true,
		Size: size,
	}
}
func (sc *ScrollContext) Encode() (string, error) {
	b, err := json.Marshal(sc)
	if err != nil {
		return "", fmt.Errorf("marshal-json is failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (sc *ScrollContext) Decode(scrollID string) error {
	b, err := base64.StdEncoding.DecodeString(scrollID)
	if err != nil {
		return fmt.Errorf("decode base64 is failed: %w", err)
	}
	if err := json.Unmarshal(b, sc); err != nil {
		return fmt.Errorf("unmarshal-json is failed: %w", err)
	}
	return nil
}

func (sc *ScrollContext) NextState(ob interface{}) (*ScrollState, error) {
	rv := reflect.ValueOf(ob)
	rt := rv.Type()
	if rt.Kind() != reflect.Slice {
		return nil, fmt.Errorf("invalid type: %v", rt)
	}

	size := sc.Size
	n := rv.Len()
	more := n == size+1
	var latestID *ScrollT

	if more {
		rv = rv.Slice(0, n-1)
		latestIDValue := coerceScrollT(rv.Index(n - 2).FieldByName(sc.Key))
		latestID = &latestIDValue
	}
	newSC := &ScrollContext{
		Key:      sc.Key,
		OrderBy:  sc.OrderBy,
		LatestID: latestID,
		More:     more,
		Size:     size,
	}
	scrollID, err := newSC.Encode()
	if err != nil {
		return nil, fmt.Errorf("invalid value: %w", err)
	}

	return &ScrollState{
		More:     more,
		ScrollID: scrollID,
		Size:     size,
	}, nil
}

type ScrollState struct {
	More     bool   `json:"more"`
	Size     int    `json:"size"`
	ScrollID string `json:"scrollId"`
}

func (s *ScrollState) DecodeContext() (*ScrollContext, error) {
	var sc ScrollContext
	if err := sc.Decode(s.ScrollID); err != nil {
		return nil, err
	}
	return &sc, nil
}

// todo: generics?
type ScrollT = int

func coerceScrollT(v reflect.Value) int {
	return int(v.Int())
}
