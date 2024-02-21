// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package id

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

var allDigits = regexp.MustCompile(`^\d+$`)
var allHex = regexp.MustCompile(`^[a-f0-9]+$`)

func TestUniqueId(t *testing.T) {
	split := func(rest string) (timestamp, increment string) {
		return rest[:18], rest[18:]
	}

	iterations := 10000
	ids := make(map[string]struct{})
	var id, lastId string
	for i := 0; i < iterations; i++ {
		id = UniqueId()

		if _, ok := ids[id]; ok {
			t.Fatalf("Got duplicated id! %s", id)
		}

		if !strings.HasPrefix(id, UniqueIdPrefix) {
			t.Fatalf("Unique ID didn't have terraform- prefix! %s", id)
		}

		rest := strings.TrimPrefix(id, UniqueIdPrefix)

		if len(rest) != UniqueIDSuffixLength {
			t.Fatalf("PrefixedUniqueId is out of sync with UniqueIDSuffixLength, post-prefix part has wrong length! %s", rest)
		}

		timestamp, increment := split(rest)

		if !allDigits.MatchString(timestamp) {
			t.Fatalf("Timestamp not all digits! %s", timestamp)
		}

		if !allHex.MatchString(increment) {
			t.Fatalf("Increment part not all hex! %s", increment)
		}

		if lastId != "" && lastId >= id {
			t.Fatalf("IDs not ordered! %s vs %s", lastId, id)
		}

		ids[id] = struct{}{}
		lastId = id
	}

	id1 := UniqueId()
	time.Sleep(time.Millisecond)
	id2 := UniqueId()
	timestamp1, _ := split(strings.TrimPrefix(id1, UniqueIdPrefix))
	timestamp2, _ := split(strings.TrimPrefix(id2, UniqueIdPrefix))

	if timestamp1 == timestamp2 {
		t.Fatalf("Timestamp part should update at least once a millisecond %s %s",
			id1, id2)
	}
}
