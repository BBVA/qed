package publish

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bbva/qed/balloon"
)

type publisher interface {
	Publish()
}

var client = &http.Client{}

func SpawnPublishers(ch <-chan balloon.Commitment) {
	numGoroutines := 10
	for i := 0; i < numGoroutines; i++ {
		go Publish(ch)
	}
}

func Publish(ch <-chan balloon.Commitment) {

	for {
		commitment := <-ch
		key := strconv.FormatUint(commitment.Version, 10)
		value := "test"

		req, _ := http.NewRequest("GET", "http://127.0.0.1:4001/gossip/add", nil)
		q := req.URL.Query()
		q.Add("key", key)
		q.Add("val", value)
		req.URL.RawQuery = q.Encode()

		rs, err := client.Do(req)
		if err != nil {
			fmt.Println("Errored when sending request to the server")
			return
		}
		defer rs.Body.Close()
	}
}
