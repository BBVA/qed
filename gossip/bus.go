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
	"sync"

	"github.com/bbva/qed/log"
)

// A Subscriber to agent gossip message queues
// is like:
//	func (my *BatchProcessor) Subscribe(c chan *Message) MessageType {
//		my.inCh = c
//		return BatchMessageType
//	}
// and receives a channel to read publised messages from
// the selected MessageType
type Subscriber interface {
	Subscribe(id int, ch <-chan *Message)
}

// A producers fills the chan *Message with the
// gossip messages to be consumed by a subscriber
type Producer interface {
	Produce(ch chan<- *Message)
}

type Subscribers []chan *Message

// Implements a subscriber / publisher model for
// gossip Messages
type MessageBus struct {
	pool [MAXMESSAGEID]Subscribers
	log  log.Logger
	rm   sync.RWMutex
}

// Publish a message to all the subscribers of its MessageType.
//
// If there is no subscriber the message will not be sent and will be lost.
//
// All the subscribers will get all the messages. If a subscriber is busy
// it will block delivery to the next subscribers. Also
// publish will create a goroutine per message sent, and
// will not time out.
//
func (eb *MessageBus) Publish(msg *Message) error {
	eb.rm.RLock()
	defer eb.rm.RUnlock()
	if chans := eb.pool[msg.Kind]; len(chans) > 0 {
		eb.log.Debugf("Agent message bus publising message to %d subscribers", len(chans))
		channels := append(chans[:0:0], chans...)
		go func(msg *Message, subscribers Subscribers) {
			for _, s := range subscribers {
				s <- msg
			}
		}(msg, channels)
		return nil
	}
	eb.log.Debugf("Agent message bus publising message: no subscribers for message kind %d ", msg.Kind)
	return NoSubscribersFound
}

// Subscribe add a subscriber to the its correspondant pool.
// Returns the subscription id needed for unsubscribe
func (eb *MessageBus) Subscribe(t MessageType, s Subscriber, size int) {
	eb.rm.Lock()
	defer eb.rm.Unlock()
	ch := make(chan *Message, size)

	if eb.pool[t] != nil {
		eb.pool[t] = append(eb.pool[t], ch)
	} else {
		eb.pool[t] = append(Subscribers{}, ch)
	}

	s.Subscribe(len(eb.pool[t]), ch)
}

// Unsubscribe a subscriber by its id
func (eb *MessageBus) Unsubscribe(t MessageType, id int) {
	eb.rm.Lock()
	eb.pool[t] = append(eb.pool[t][:id], eb.pool[t][id+1:]...)
	eb.rm.Unlock()
}

// Implements a message queue in which
// subscribers consumes producers
// Messages.
//
// There is a queue for each kind of message,
// and all the producers and subscribers
// will operate over the same chan *Message.
//
// This pattern allows a pool of subscribers to
// consume messages from a pool of producers
// without blocking.
type MessageQueue struct {
	size  int
	queue [MAXMESSAGEID]chan *Message
	rm    sync.RWMutex
}

// Register a producer to the MessageType queue
func (mq *MessageQueue) Producer(t MessageType, p Producer) {
	mq.rm.RLock()
	defer mq.rm.RUnlock()

	if mq.queue[t] == nil {
		mq.queue[t] = make(chan *Message, mq.size)
	}
	p.Produce(mq.queue[t])
}

// Register a consumer to the MessageType queue
func (mq *MessageQueue) Consumer(t MessageType, s Subscriber) {
	mq.rm.RLock()
	defer mq.rm.RUnlock()
	if mq.queue[t] == nil {
		mq.queue[t] = make(chan *Message, mq.size)
	}
	s.Subscribe(0, mq.queue[t])
}

// Cancels signals all producers and consumers to stop
// closing the internal channel
func (mq *MessageQueue) Cancel(t MessageType) {
	mq.rm.RLock()
	defer mq.rm.RUnlock()
	close(mq.queue[t])
}
