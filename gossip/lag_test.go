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
package gossip

import (
	"testing"
	"time"

	"github.com/bbva/qed/protocol"
	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	c := NewCounter()
	require.Equal(t, c.Get(), uint64(0), "Counter value must be 0 when initialized")
	c.Add(1)
	require.Equal(t, c.Get(), uint64(1), "Counter value must be 1 when added 1 in a new counter")
	c.Set(10)
	require.Equal(t, c.Get(), uint64(10), "Counter value must be 10 when set to 10")
}

func TestCounterRate(t *testing.T) {
	c := NewCounter()
	c.Start(1 * time.Second)
	c.Add(100)
	time.Sleep(1100 * time.Millisecond)
	require.Equal(t, uint64(100), c.Rate(), "Rate value must be 100 when adding 100 in one second interval")
	c.Stop()
}

func TestNoLag(t *testing.T) {
	l := NewLag()

	newSignedSnapshot := func(v uint64) *protocol.SignedSnapshot {
		return &protocol.SignedSnapshot{
			Snapshot: &protocol.Snapshot{
				Version: v,
			},
		}
	}

	newBatchSnapshots := func(n uint64) *protocol.BatchSnapshots {
		var snaps []*protocol.SignedSnapshot
		for i := uint64(0); i < n; i++ {
			snaps = append(snaps, newSignedSnapshot(i))
		}
		return &protocol.BatchSnapshots{
			Snapshots: snaps,
		}
	}

	batch := newBatchSnapshots(100)
	l.Start(1 * time.Second)
	l.Process(batch)
	time.Sleep(1100 * time.Millisecond)
	l.Stop()
	require.Equal(t, uint64(0), l.Lag.Get(), "There should be no lag")

}

func TestLag(t *testing.T) {

	l := NewLag()

	newSignedSnapshot := func(v uint64) *protocol.SignedSnapshot {
		return &protocol.SignedSnapshot{
			Snapshot: &protocol.Snapshot{
				Version: v,
			},
		}
	}

	newBatchSnapshots := func(offset, n uint64) *protocol.BatchSnapshots {
		var snaps []*protocol.SignedSnapshot
		for i := offset; i < n; i++ {
			snaps = append(snaps, newSignedSnapshot(i))
		}
		return &protocol.BatchSnapshots{
			Snapshots: snaps,
		}
	}

	batch1 := newBatchSnapshots(0, 100)
	batch2 := newBatchSnapshots(100, 200)
	batch3 := newBatchSnapshots(200, 300)
	l.Start(1 * time.Second)
	l.Process(batch1)
	l.Process(batch2)
	l.Process(batch3)
	time.Sleep(1100 * time.Millisecond)
	l.Process(batch1)
	l.Stop()
	require.Equal(t, uint64(200), l.Lag.Get(), "There should be a lag of 200 snapshots")

}
