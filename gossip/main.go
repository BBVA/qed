package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/bbva/qed/gossip/api"
	"github.com/bbva/qed/gossip/backends"
	"github.com/bbva/qed/gossip/gossip"
)

var (
	members = flag.String("members", "", "comma seperated list of members")
	port    = flag.Int("port", 4001, "http port")
)

func init() {
	flag.Parse()
}

func main() {

	ctx := gossip.Context{
		Mtx:   sync.RWMutex{},
		Items: map[string]string{},
	}

	if err := gossip.StartGossip(ctx, members); err != nil {
		fmt.Println(err)
	}

	client := backends.NewRedisClient()

	http.HandleFunc("/gossip/add", func(w http.ResponseWriter, r *http.Request) { api.AddHandler(ctx, w, r, client) })
	http.HandleFunc("/gossip/del", func(w http.ResponseWriter, r *http.Request) { api.DelHandler(ctx, w, r) })
	http.HandleFunc("/gossip/get", func(w http.ResponseWriter, r *http.Request) { api.GetHandler(ctx, w, r) })
	fmt.Printf("Listening on :%d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Println(err)
	}
}
