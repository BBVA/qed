package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func main() {
	conn := flag.Int("c", 1, "Max open idle connections")
	workers := flag.Uint64("w", 1, "Request per second")
	timeout := flag.Duration("t", 30*time.Second, "Timeout")
	duration := flag.Duration("d", 10*time.Second, "Duration")
	endpoint := flag.String("e", "http://localhost:8080/events", "Endpoint")
	apikey := flag.String("k", "apikey", "apikey")
	rate := flag.Uint64("r", 100, "Request per second")
	flag.Parse()

	targeter := myTargeter(*endpoint, http.Header{"Api-Key": []string{*apikey}})

	atk := vegeta.NewAttacker(vegeta.Connections(*conn), vegeta.Workers(*workers), vegeta.Timeout(*timeout))
	res := atk.Attack(targeter, *rate, *duration)
	enc := vegeta.NewEncoder(os.Stdout)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for {
		select {
		case <-sig:
			atk.Stop()
			os.Exit(0)
		case r, ok := <-res:
			if !ok {
				os.Exit(-1)
			}
			if err := enc.Encode(r); err != nil {
				os.Exit(-1)
			}
		}
	}

}

func myTargeter(endpoint string, hdr http.Header) vegeta.Targeter {
	var mu sync.Mutex

	return func(tgt *vegeta.Target) (err error) {
		mu.Lock()
		defer mu.Unlock()

		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		tgt.Body = []byte(fmt.Sprintf(`{"event": "%s"}`, []byte(base64.StdEncoding.EncodeToString([]byte(time.Now().String())))))
		tgt.Header = hdr
		tgt.Method = "POST"
		tgt.URL = endpoint

		return nil
	}
}
