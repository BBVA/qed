package main

// import (
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"sync"

// 	"github.com/bbva/qed/gossip/backends"
// 	"github.com/hashicorp/memberlist"
// 	"github.com/pborman/uuid"
// )

// var (
// 	mtx        sync.RWMutex
// 	members    = flag.String("members", "", "comma seperated list of members")
// 	port       = flag.Int("port", 4001, "http port")
// 	items      = map[string]string{}
// 	broadcasts *memberlist.TransmitLimitedQueue
// )

// type broadcast struct {
// 	msg    []byte
// 	notify chan<- struct{}
// }

// type delegate struct{}

// type update struct {
// 	Action string // add, del
// 	Data   map[string]string
// }

// func init() {
// 	flag.Parse()
// }

// func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
// 	return false
// }

// func (b *broadcast) Message() []byte {
// 	return b.msg
// }

// func (b *broadcast) Finished() {
// 	if b.notify != nil {
// 		close(b.notify)
// 	}
// }

// func (d *delegate) NodeMeta(limit int) []byte {
// 	return []byte{}
// }

// func (d *delegate) NotifyMsg(b []byte) {
// 	if len(b) == 0 {
// 		return
// 	}

// 	switch b[0] {
// 	case 'd': // data
// 		var updates []*update
// 		if err := json.Unmarshal(b[1:], &updates); err != nil {
// 			return
// 		}
// 		mtx.Lock()
// 		for _, u := range updates {
// 			for k, v := range u.Data {
// 				switch u.Action {
// 				case "add":
// 					items[k] = v
// 				case "del":
// 					delete(items, k)
// 				}
// 			}
// 		}
// 		mtx.Unlock()
// 	}
// }

// func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
// 	return broadcasts.GetBroadcasts(overhead, limit)
// }

// func (d *delegate) LocalState(join bool) []byte {
// 	mtx.RLock()
// 	m := items
// 	mtx.RUnlock()
// 	b, _ := json.Marshal(m)
// 	return b
// }

// func (d *delegate) MergeRemoteState(buf []byte, join bool) {
// 	if len(buf) == 0 {
// 		return
// 	}
// 	if !join {
// 		return
// 	}
// 	var m map[string]string
// 	if err := json.Unmarshal(buf, &m); err != nil {
// 		return
// 	}
// 	mtx.Lock()
// 	for k, v := range m {
// 		items[k] = v
// 	}
// 	mtx.Unlock()
// }

// func gossipBroadcast(w http.ResponseWriter, action, key string) error {
// 	b, err := json.Marshal([]*update{
// 		&update{
// 			Action: action,
// 			Data: map[string]string{
// 				key: key,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return err
// 	}

// 	broadcasts.QueueBroadcast(&broadcast{
// 		msg:    append([]byte("d"), b...),
// 		notify: nil,
// 	})

// 	return nil
// }

// func addHandler(w http.ResponseWriter, r *http.Request, client *backends.Client) {
// 	r.ParseForm()
// 	key := r.Form.Get("key")
// 	val := r.Form.Get("val")
// 	mtx.Lock()
// 	items[key] = val // fatal error: concurrent map read and map write
// 	go client.Publish(key, val)
// 	mtx.Unlock()

// 	gossipBroadcast(w, "add", key)
// }

// func delHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	key := r.Form.Get("key")
// 	mtx.Lock()
// 	delete(items, key)
// 	mtx.Unlock()

// 	gossipBroadcast(w, "del", "")
// }

// func getHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	key := r.Form.Get("key")
// 	mtx.RLock()
// 	val := items[key]
// 	mtx.RUnlock()
// 	w.Write([]byte(val))
// }

// func startGossip() error {
// 	hostname, _ := os.Hostname()
// 	c := memberlist.DefaultLocalConfig()
// 	c.Delegate = &delegate{}
// 	c.BindPort = 0
// 	c.Name = hostname + "-" + uuid.NewUUID().String()
// 	m, err := memberlist.Create(c)
// 	if err != nil {
// 		return err
// 	}
// 	if len(*members) > 0 {
// 		parts := strings.Split(*members, ",")
// 		_, err := m.Join(parts)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	broadcasts = &memberlist.TransmitLimitedQueue{
// 		NumNodes: func() int {
// 			return m.NumMembers()
// 		},
// 		RetransmitMult: 3,
// 	}
// 	node := m.LocalNode()
// 	fmt.Printf("Local member %s:%d\n", node.Addr, node.Port)
// 	return nil
// }

// func main() {
// 	if err := startGossip(); err != nil {
// 		fmt.Println(err)
// 	}

// 	client := backends.NewRedisClient()

// 	http.HandleFunc("/gossip/add", func(w http.ResponseWriter, r *http.Request) { addHandler(w, r, client) })
// 	http.HandleFunc("/gossip/del", delHandler)
// 	http.HandleFunc("/gossip/get", getHandler)
// 	fmt.Printf("Listening on :%d\n", *port)
// 	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
// 		fmt.Println(err)
// 	}
// }
