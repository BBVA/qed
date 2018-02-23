package sequencer

import (
	"fmt"
	"testing"
	"verifiabledata/api"
	"verifiabledata/util"
)

func TestSingleInsertionSequencing(t *testing.T) {

	data := api.InsertData{Message: "Event 1"}
	request := api.InsertRequest{
		InsertData:      data,
		ResponseChannel: make(chan *api.InsertResponse),
	}

	sequencer := NewSequencer(10)
	sequencer.Start()

	sequencer.InsertRequestQueue <- &request

	response := <-request.ResponseChannel
	fmt.Println(response)

	if response.Index != 0 {
		t.Errorf("The index should be 0")
	}

	if response.Hash != string(util.Hash([]byte(data.Message))) {
		t.Errorf("The hash of the original message doesn't match the returned hash")
	}

	if response.Commitment != string(util.Hash([]byte(data.Message))) { // TODO CHANGE THIS!!
		t.Errorf("The commitment doesn't matches the hash of the original message")
	}

	sequencer.Stop()

}
