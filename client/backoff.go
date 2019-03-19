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
	"math"
	"math/rand"
	"sync"
	"time"
)

// BackoffF specifies the signature of a function that returns the
// time to wait before the next call to a resource. To stop retrying
// return false in the 2nd return value.
type BackoffF func(attempt int) (time.Duration, bool)

// Backoff allows callers to implement their own Backoff strategy.
type Backoff interface {
	// Next implements a BackoffF.
	Next(attempt int) (time.Duration, bool)
}

// StopBackoff is a fixed backoff policy that always returns false for
// Next(), meaning that the operation should never be retried.
type StopBackoff struct{}

// NewStopBackoff returns a new StopBackoff.
func NewStopBackoff() *StopBackoff {
	return &StopBackoff{}
}

// Next implements BackoffF for StopBackoff.
func (b StopBackoff) Next(attempt int) (time.Duration, bool) {
	return 0, false
}

// ConstantBackoff is a backoff policy that always returns the same delay.
type ConstantBackoff struct {
	interval time.Duration
}

// NewConstantBackoff returns a new ConstantBackoff.
func NewConstantBackoff(interval time.Duration) *ConstantBackoff {
	return &ConstantBackoff{interval: interval}
}

// Next implements BackoffF for ConstantBackoff.
func (b *ConstantBackoff) Next(attempt int) (time.Duration, bool) {
	return b.interval, true
}

// SimpleBackoff takes a list of fixed values for backoff intervals.
// Each call to Next returns the next value from that fixed list.
// After each value is returned, subsequent calls to Next will only return
// the last element.
type SimpleBackoff struct {
	sync.Mutex
	ticks []int
}

// NewSimpleBackoff creates a SimpleBackoff algorithm with the specified
// list of fixed intervals in milliseconds.
func NewSimpleBackoff(ticks ...int) *SimpleBackoff {
	return &SimpleBackoff{ticks: ticks}
}

// Next implements BackoffF for SimpleBackoff.
func (b *SimpleBackoff) Next(attempt int) (time.Duration, bool) {
	b.Lock()
	defer b.Unlock()
	if attempt >= len(b.ticks) {
		return 0, false
	}
	ms := b.ticks[attempt]
	return time.Duration(ms) * time.Millisecond, true
}

// ExponentialBackoff implements the simple exponential backoff described by
// Douglas Thain at http://dthain.blogspot.de/2009/02/exponential-backoff-in-distributed.html.
type ExponentialBackoff struct {
	t float64 // initial timeout (in msec)
	f float64 // exponential factor (e.g. 2)
	m float64 // maximum timeout (in msec)
}

// NewExponentialBackoff returns a ExponentialBackoff backoff policy.
// Use initialTimeout to set the first/minimal interval
// and maxTimeout to set the maximum wait interval.
func NewExponentialBackoff(initialTimeout, maxTimeout time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		t: float64(int64(initialTimeout / time.Millisecond)),
		f: 2.0,
		m: float64(int64(maxTimeout / time.Millisecond)),
	}
}

// Next implements BackoffF for ExponentialBackoff.
func (b *ExponentialBackoff) Next(attempt int) (time.Duration, bool) {
	r := 1.0 + rand.Float64() // random number in [1..2]
	m := math.Min(r*b.t*math.Pow(b.f, float64(attempt)), b.m)
	if m >= b.m {
		return 0, false
	}
	d := time.Duration(int64(m)) * time.Millisecond
	return d, true
}
