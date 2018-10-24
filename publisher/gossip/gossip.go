package gossip

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
)

var (
	broadcasts *memberlist.TransmitLimitedQueue
)

type Snapshot struct {
	Version    string
	Commitment string
}
type Context struct {
	Mtx       sync.RWMutex
	Snapshots []Snapshot
	// Items map[string]string
}

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct {
	mtx       sync.RWMutex
	snapshots []Snapshot
	// items map[string]string
}

type update struct {
	Action string // add, del
	Data   map[string]string
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd': // data
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		d.mtx.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				snapshot := Snapshot{Version: k, Commitment: v}
				switch u.Action {
				case "add":
					d.snapshots = append(d.snapshots, snapshot)
					// d.items[k] = v
					// case "del":
					// 	delete(d.items, k)
				}
			}
		}
		d.mtx.Unlock()
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	d.mtx.RLock()
	// m := d.items
	m := d.snapshots
	d.mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	d.mtx.Lock()
	for k, v := range m {
		snapshot := Snapshot{Version: k, Commitment: v}
		d.snapshots = append(d.snapshots, snapshot)
		// d.items[k] = v
	}
	d.mtx.Unlock()
}

func GossipBroadcast(action, key string) error {
	b, err := json.Marshal([]*update{
		&update{
			Action: action,
			Data: map[string]string{
				key: key,
			},
		},
	})

	if err != nil {
		// http.Error(w, err.Error(), 500)
		return err
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})

	return nil
}

func StartGossip(ctx *Context, members *string) error {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLocalConfig()
	c.Delegate = &delegate{
		mtx:       ctx.Mtx,
		snapshots: ctx.Snapshots,
	}
	c.BindPort = 0
	c.Name = hostname + "-" + uuid.NewUUID().String()
	m, err := memberlist.Create(c)
	if err != nil {
		return err
	}
	if len(*members) > 0 {
		parts := strings.Split(*members, ",")
		_, err := m.Join(parts)
		if err != nil {
			return err
		}
	}
	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	node := m.LocalNode()
	fmt.Printf("Local member %s:%d\n", node.Addr, node.Port)
	return nil
}
