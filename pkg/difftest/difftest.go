package difftest

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"unsafe"

	"github.com/gookit/color"
	"github.com/shibukawa/cdiff"
)

// TODO: write to files
func AssertGotAndWantBytes(t *testing.T, got interface{}, want []byte) {
	t.Helper()

	// normalize
	var wantString, gotString string
	{
		b, err := json.Marshal(got)
		if err != nil {
			t.Fatalf("marshal for got, %+v, %+v", got, err)
		}
		var v interface{}
		if err := json.Unmarshal(b, &v); err != nil {
			t.Fatalf("marshal,unmarshal for got, %+v, %+v", got, err)
		}
		b2, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("marshal,unmarshal,marshal for got, %+v, %+v", got, err)
		}
		gotString = strings.TrimSpace(*(*string)(unsafe.Pointer(&b2)))
	}
	{
		var v interface{}
		if err := json.Unmarshal(want, &v); err != nil {
			t.Fatalf("unmarshal for want, %+v, %+v", string(want), err)
		}
		b2, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatalf("unmarshal,marshal for want, %+v, %+v", string(want), err)
		}
		wantString = strings.TrimSpace(*(*string)(unsafe.Pointer(&b2)))
	}

	diff := cdiff.Diff(gotString, wantString, cdiff.WordByWord)
	output := diff.UnifiedWithGooKitColor("got", "want", 3, cdiff.GooKitColorTheme)

	// first 4lines is header
	if len(strings.SplitAfterN(output, "\n", 5)) >= 5 {
		t.Errorf(color.Sprintf(output))
		if testing.Verbose() {
			t.Logf("want\n%s", wantString)
			t.Logf("but got\n%s", gotString)
		}
	}
}

func AssertGotAndWantString(t *testing.T, got interface{}, want string) {
	t.Helper()
	AssertGotAndWantBytes(t, got, []byte(want))
}

func LogDiffGotStringAndWantString(t *testing.T, gotString, wantString string) {
	t.Helper()

	diff := cdiff.Diff(gotString, wantString, cdiff.WordByWord)
	output := diff.UnifiedWithGooKitColor("got", "want", 3, cdiff.GooKitColorTheme)

	// first 4lines is header
	t.Errorf(color.Sprintf(output))
	if testing.Verbose() {
		t.Logf("want\n%s", wantString)
		t.Logf("but got\n%s", gotString)
	}
}

func NormalizeString(s string) string {
	return strings.TrimSpace(spaceonlyRX.ReplaceAllString(s, "\n"))
}

var spaceonlyRX = regexp.MustCompile(`(?m)^\s+`)
