package sequencer

import (
	"fmt"
	"sync"
	"testing"
	api "verifiabledata/api/http"
	"verifiabledata/util"
)

func TestSingleInsertionSequencing(t *testing.T) {

	data := api.InsertData{Message: "Event 1"}
	request := &api.InsertRequest{
		InsertData:      data,
		ResponseChannel: make(chan *api.InsertResponse),
	}

	sequencer := NewSequencer(10)
	sequencer.Start()

	sequencer.Enqueue(request)

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

func TestMultipleInsertionSequencing(t *testing.T) {

	sequencer := NewSequencer(10)
	sequencer.Start()

	var wg sync.WaitGroup
	wg.Add(10)

	requests := make([]*api.InsertRequest, 10)
	for i := 0; i < 10; i++ {
		data := api.InsertData{Message: fmt.Sprintf("Event %d", i)}
		requests[i] = &api.InsertRequest{
			InsertData:      data,
			ResponseChannel: make(chan *api.InsertResponse),
		}
	}

	for i, req := range requests {
		go func(index int, request *api.InsertRequest) {
			response := <-request.ResponseChannel
			if response.Index != uint64(index) {
				t.Errorf("The assigned index doesn't obey the insertion order")
			}
			wg.Done()
		}(i, req)
		sequencer.Enqueue(req)
	}

	wg.Wait()
	sequencer.Stop()

}

func BenchmarkSingleInsertion(b *testing.B) {

	data := api.InsertData{Message: "Event 1"}
	request := &api.InsertRequest{
		InsertData:      data,
		ResponseChannel: make(chan *api.InsertResponse),
	}

	sequencer := NewSequencer(10)
	sequencer.Start()

	for i := 0; i < b.N; i++ {
		sequencer.Enqueue(request)
		<-request.ResponseChannel
	}

	sequencer.Stop()
}
