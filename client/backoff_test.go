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

package client

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStopBackoff(t *testing.T) {
	b := NewStopBackoff()
	_, ok := b.Next(0)
	require.False(t, ok)
}

func TestConstantBackoff(t *testing.T) {
	b := NewConstantBackoff(time.Second)
	d, ok := b.Next(0)
	require.True(t, ok)
	require.Equal(t, time.Second, d)
}

func TestSimpleBackoff(t *testing.T) {

	testCases := []struct {
		Duration time.Duration
		Continue bool
	}{
		{
			Duration: 1 * time.Millisecond,
			Continue: true,
		},
		{
			Duration: 2 * time.Millisecond,
			Continue: true,
		},
		{
			Duration: 7 * time.Millisecond,
			Continue: true,
		},
		{
			Duration: 0,
			Continue: false,
		},
		{
			Duration: 0,
			Continue: false,
		},
	}

	b := NewSimpleBackoff(1, 2, 7)

	for i, c := range testCases {
		d, ok := b.Next(i)
		require.Equalf(t, c.Continue, ok, "The continue value should match for test case %d", i)
		require.Equalf(t, c.Duration, d, "The duration value should match for test case %d", i)
	}
}

func TestExponentialBackoff(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	min := time.Duration(8) * time.Millisecond
	max := time.Duration(256) * time.Millisecond
	b := NewExponentialBackoff(min, max)

	between := func(value time.Duration, a, b int) bool {
		x := int(value / time.Millisecond)
		return a <= x && x <= b
	}

	d, ok := b.Next(0)
	require.True(t, ok)
	require.True(t, between(d, 8, 256))

	d, ok = b.Next(1)
	require.True(t, ok)
	require.True(t, between(d, 8, 256))

	d, ok = b.Next(3)
	require.True(t, ok)
	require.True(t, between(d, 8, 256))

	d, ok = b.Next(4)
	require.True(t, ok)
	require.True(t, between(d, 8, 256))

	_, ok = b.Next(5)
	require.False(t, ok)

	_, ok = b.Next(6)
	require.False(t, ok)

}
