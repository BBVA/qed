package publish

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bbva/qed/balloon"
)

type publisher interface {
	Publish()
}

var client = &http.Client{}
var url = "http://127.0.0.1"
var port = "4001"
var endpoint = "/gossip/add"

func SpawnPublishers(ch <-chan *balloon.Commitment) {
	numGoRoutines := 100
	for i := 0; i < numGoRoutines; i++ {
		go func() {
			for {
				commitment := <-ch
				Publish(commitment)
			}
		}()
	}
}

func Publish(commitment *balloon.Commitment) {
	key := strconv.FormatUint(commitment.Version, 10)
	value := "test"

	req, _ := http.NewRequest("GET", url+":"+port+endpoint, nil)
	q := req.URL.Query()
	q.Add("key", key)
	q.Add("val", value)
	req.URL.RawQuery = q.Encode()

	rs, err := client.Do(req)
	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return
	}
	io.Copy(ioutil.Discard, rs.Body)
	defer rs.Body.Close()
}
