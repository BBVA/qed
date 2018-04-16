// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package agent

import (
	"testing"
)

func TestAdd(t *testing.T) {

	testAgent, _ := NewAgent("http://localhost:8080")
	testAgent.Add("Hola mundo!")

}

func BenchmarkAdd(b *testing.B) {

	testAgent, _ := NewAgent("http://localhost:8080")

	for n := 0; n < b.N; n++ {
		testAgent.Add(string(n))
	}

}
