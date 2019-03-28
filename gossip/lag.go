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
	"sync/atomic"
	"time"

	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

type Counter struct {
	n    uint64
	rate uint64
	quit chan bool
}

func NewCounter() *Counter {
	return &Counter{
		quit: make(chan bool),
	}
}

func (c *Counter) Start(interval time.Duration) {

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r := float64(c.n) / interval.Seconds()
				atomic.StoreUint64(&c.rate, uint64(r))
			case <-c.quit:
				return
			}
		}
	}()
}

func (c *Counter) Stop() {
	select {
	case <-c.quit:
		return
	default:
	}
	close(c.quit)
}

func (c *Counter) Add(val uint64) {
	atomic.AddUint64(&c.n, val)
}

func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.n)
}

func (c *Counter) Set(val uint64) {
	atomic.StoreUint64(&c.n, val)
}

func (c *Counter) Rate() uint64 {
	return atomic.LoadUint64(&c.rate)
}

// The lag processor measures the difference  between the expeted
// version given the actual throughput against versions present in batches
type Lag struct {
	Batches   *Counter
	Snapshots *Counter
	Version   *Counter
	Lag       *Counter
}

func NewLag() *Lag {
	return &Lag{
		Batches:   NewCounter(),
		Snapshots: NewCounter(),
		Version:   NewCounter(),
		Lag:       NewCounter(),
	}
}

func (l Lag) Start(d time.Duration) {
	l.Batches.Start(d)
	l.Snapshots.Start(d)
}

func (l Lag) Stop() {
	l.Batches.Stop()
	l.Snapshots.Stop()
}

func (l Lag) Get() uint64 {
	return l.Lag.Get()
}

func (l Lag) RegisterMetrics(srv *metrics.Server) {
	metrics := []prometheus.Collector{}

	for _, m := range metrics {
		srv.Register(m)
	}
}

func (l *Lag) Process(b *protocol.BatchSnapshots) {
	l.Batches.Add(1)
	l.Snapshots.Add(uint64(len(b.Snapshots)))

	last := l.Version.Get()
	curr := b.Snapshots[len(b.Snapshots)-1].Snapshot.Version

	if curr > last {
		l.Lag.Set(0)
		l.Version.Set(curr)
		return
	}

	l.Lag.Set(last - curr)
}

