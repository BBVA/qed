package api

import (
	"net/http"

	"github.com/bbva/qed/publisher/backends"
	"github.com/bbva/qed/publisher/gossip"
)

func AddHandler(ctx gossip.Context, w http.ResponseWriter, r *http.Request, client *backends.Client) {
	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("val")
	snapshot := gossip.Snapshot{Version: key, Commitment: val}
	ctx.Mtx.Lock()
	ctx.Snapshots = append(ctx.Snapshots, snapshot)
	//ctx.Items[key] = val // fatal error: concurrent map read and map write
	ctx.Mtx.Unlock()
	go client.Publish(key, val)

	gossip.GossipBroadcast("add", key)
}

// func DelHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	key := r.Form.Get("key")
// 	ctx.Mtx.Lock()
// 	delete(ctx.Items, key)
// 	ctx.Mtx.Unlock()

// 	gossip.GossipBroadcast("del", "")
// }

// func GetHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	key := r.Form.Get("key")
// 	ctx.Mtx.RLock()
// 	val := ctx.Items[key]
// 	ctx.Mtx.RUnlock()
// 	w.Write([]byte(val))
// }
