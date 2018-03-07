// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package util

import (
	"fmt"
	"testing"
)

func TestHash256(t *testing.T) {
	expected_hash := "6d1103674f29502c873de14e48e9e432ec6cf6db76272c7b0dad186bb92c9a9a"
	buff := []byte("Hello World1")
	h := Hash256()
	d := h(buff)

	if expected_hash != fmt.Sprintf("%x", d) {
		t.Fatal("Unexpected Hash when hasing buffer")
	}
}
