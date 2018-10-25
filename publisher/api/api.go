package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bbva/qed/publisher/backends"
	"github.com/bbva/qed/publisher/gossip"
)

func AddHandler(ctx gossip.Context, w http.ResponseWriter, r *http.Request, client *backends.Client) {

	if r.Method == "POST" {

		var ss gossip.SignedSnapshot
		err := json.NewDecoder(r.Body).Decode(&ss)
		if err != nil {
			fmt.Println("Error unmarshalling: ", err)
		}

		ctx.Mtx.Lock()
		ctx.Snapshots = append(ctx.Snapshots, ss)
		ctx.Mtx.Unlock()

		key := strconv.FormatUint(ss.Snapshot.Version, 10)
		v := sha256.Sum256(ss.Snapshot.HistoryDigest)
		val := string(v[:])
		go client.Publish(key, val)

		gossip.GossipBroadcast("add", key)

	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

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
