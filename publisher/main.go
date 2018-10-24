package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/bbva/qed/publisher/api"
	"github.com/bbva/qed/publisher/backends"
	"github.com/bbva/qed/publisher/gossip"
)

var (
	members string
	port    int
)

func init() {
	flag.StringVar(&members, "members", "", "comma seperated list of members")
	flag.IntVar(&port, "port", 4001, "http port")
}

func main() {
	flag.Parse()
	ctx := gossip.Context{
		Mtx:       sync.RWMutex{},
		Snapshots: []gossip.Snapshot{},
		//		Items: map[string]string{},
	}

	if err := gossip.StartGossip(&ctx, &members); err != nil {
		fmt.Println(err)
	}

	client := backends.NewRedisClient()

	http.HandleFunc("/gossip/add", func(w http.ResponseWriter, r *http.Request) { api.AddHandler(ctx, w, r, client) })
	// http.HandleFunc("/gossip/del", func(w http.ResponseWriter, r *http.Request) { api.DelHandler(w, r) })
	// http.HandleFunc("/gossip/get", func(w http.ResponseWriter, r *http.Request) { api.GetHandler(w, r) })
	fmt.Printf("Listening on :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Println(err)
	}
}
