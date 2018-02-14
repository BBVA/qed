// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"testing"
	"verifiabledata/store/memory"
)

var eventTable = []struct {
	index  int
	digest string
	event  string
}{
	{0, "6d1103674f29502c873de14e48e9e432ec6cf6db76272c7b0dad186bb92c9a9a", "Hello World1"},
}

func TestAddEvent(t *testing.T) {
	s := memory.NewMemoryStore()
	ht := NewHistoryTree(s)

	for e := range eventTable {
		commitment, err := tree.AddEvent(e.event)
		if err != nil {
			t.Fatal("Error in AddEvent call: ", e, err)
		}

		if index != e.version {
			t.Fatal("Error in AddEvent call: ", index)
		}

		if e.digest != fmt.Sprintf("%x", commitment.Digest) {
			t.Fatal("Error in AddEvent call: ", commitment, c)
		}
	}
}
