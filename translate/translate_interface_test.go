package translate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/podhmo/apikit/resolve"
)

type DB struct{}
type User struct{}

func ListUser(db *DB) []User { return nil }

func TestTracker(t *testing.T) {
	resolver := resolve.NewResolver()
	tracker := NewTracker()

	def := resolver.Resolve(ListUser)
	tracker.Track(def)

	var buf strings.Builder
	WriteInterface(&buf, tracker, "Component")
	fmt.Println(buf.String())
}
