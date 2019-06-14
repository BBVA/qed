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

	"github.com/stretchr/testify/require"
)

type testSubscriber struct {
	ch <-chan *Message
	id int
}

func (ts *testSubscriber) Subscribe(id int, ch <-chan *Message) {
	ts.ch = ch
	ts.id = id
}

func TestMessageBus(t *testing.T) {
	var mb MessageBus
	var ts testSubscriber
	m1 := &Message{
		Kind:    BatchMessageType,
		From:    nil,
		TTL:     0,
		Payload: nil,
	}
	mb.Subscribe(BatchMessageType, &ts, 1)
	_ = mb.Publish(m1)
	m2 := <-ts.ch
	require.Equal(t, m2, m1, "Messages should match")
}

type testProducer struct {
	ch chan<- *Message
}

func (tp *testProducer) Produce(ch chan<- *Message) {
	tp.ch = ch
}

func TestMessageQueue(t *testing.T) {
	var mq MessageQueue
	var ts testSubscriber
	var tp testProducer
	var m2 *Message

	m1 := &Message{
		Kind:    BatchMessageType,
		From:    nil,
		TTL:     0,
		Payload: nil,
	}
	mq.Consumer(BatchMessageType, &ts)
	mq.Producer(BatchMessageType, &tp)

	go func() { m2 = <-ts.ch }()
	tp.ch <- m1

	require.Equal(t, m2, m1, "Messages should match")
}
