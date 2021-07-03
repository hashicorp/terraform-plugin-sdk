package resource

import (
	"regexp"
	"strings"
	"testing"
)

var all36 = regexp.MustCompile(`^[a-z0-9]+$`)

func TestUniqueId(t *testing.T) {
	iterations := 10000
	ids := make(map[string]struct{})
	var id string
	for i := 0; i < iterations; i++ {
		id = UniqueId()

		if _, ok := ids[id]; ok {
			t.Fatalf("Got duplicated id! %s", id)
		}

		if !strings.HasPrefix(id, UniqueIdPrefix) {
			t.Fatalf("Unique ID didn't have terraform- prefix! %s", id)
		}

		rest := strings.TrimPrefix(id, UniqueIdPrefix)

		if !all36.MatchString(rest) {
			t.Fatalf("Suffix isn't in base 36! %s", rest)
		}

		ids[id] = struct{}{}
	}
}
