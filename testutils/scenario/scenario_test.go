/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package scenario

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test recursive
func TestRecursive(t *testing.T) {
	let, report := New()
	defer func() {
		t.Logf("\n%v", report())
	}()

	let(t, "first level", func(t *testing.T) {
		fmt.Println("print 1")
		let(t, "second level 1", func(t *testing.T) {
			fmt.Println("print 2")
			let(t, "third level 1", func(t *testing.T) {
				fmt.Println("print 3")
			})
			let(t, "third level 2", func(t *testing.T) {
				fmt.Println("print 3")
			})
		})
		let(t, "second level 2", func(t *testing.T) {
			fmt.Println("print 4")
		})
	})
}

// Test feature spec
func TestFeatureSpec(t *testing.T) {
	let, report := New()
	defer func() {
		t.Logf("\n%v", report())
	}()

	feature := func(t *testing.T, desc string, fns ...TestF) {
		let(t, "Feature: "+desc, fns...)
	}
	scenario := func(t *testing.T, desc string, fns ...TestF) {
		let(t, "Scenario: "+desc, fns...)
	}
	given := func(t *testing.T, desc string, fns ...TestF) {
		let(t, "Given: "+desc, fns...)
	}
	when := func(t *testing.T, desc string, fns ...TestF) {
		let(t, "When: "+desc, fns...)
	}
	then := func(t *testing.T, desc string, fns ...TestF) {
		let(t, "Then: "+desc, fns...)
	}

	feature(t, "TV power button", func(t *testing.T) {
		var tv_power_button bool
		scenario(t, "User presses power button when TV is off", func(t *testing.T) {
			given(t, "A TV  set that is switched off", func(t *testing.T) {
				require.False(t, tv_power_button)
			})

			when(t, "The power button is pressed", func(t *testing.T) {
				tv_power_button = true
			})

			then(t, "The tv should switch on", func(t *testing.T) {
				assert.True(t, tv_power_button)
			})
		})
	})
}

// Test fun spec
func TestFunSpec(t *testing.T) {
	let, report := New()
	defer func() {
		t.Logf("\n%v", report())
	}()
	describe := let
	it := let
	ignore := func(t *testing.T, desc string, fns ...TestF) {
		t.Skip("Ignore: " + desc)
		let(t, "Ignore: "+desc, fns...)
	}

	describe(t, "A Set", func(t *testing.T) {
		var set []int
		describe(t, "(when empty)", func(t *testing.T) {
			set = make([]int, 0)
			it(t, "should have size 0", func(t *testing.T) {
				assert.Equal(t, 0, len(set), "")
			})
			ignore(t, "should be empty", func(t *testing.T) {})
		})
	})
}
