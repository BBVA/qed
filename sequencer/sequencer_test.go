package sequencer

import (
	"fmt"
	"sync"
	"testing"

	"verifiabledata/api/http"
	"verifiabledata/util"
)

func TestSingleInsertionSequencing(t *testing.T) {

	hasher := util.Hash256()
	data := http.InsertData{Message: "Event 1"}
	request := &http.InsertRequest{
		InsertData:      data,
		ResponseChannel: make(chan *http.InsertResponse),
	}

	sequencer := NewSequencer(10)
	sequencer.Start()

	sequencer.Enqueue(request)

	response := <-request.ResponseChannel
	fmt.Println(response)

	if response.Index != 0 {
		t.Errorf("The index should be 0")
	}

	if response.Hash != string(hasher.Do([]byte(data.Message))) {
		t.Errorf("The hash of the original message doesn't match the returned hash")
	}

	// FIXME: test the commitment in this level requires a verification process
	// if response.Commitment != string(`FILL WITH VERIFICATION COMMITMENT`) {
	// 	t.Errorf("The commitment doesn't matches the hash of the original message")
	// }

	sequencer.Stop()

}

func TestMultipleInsertionSequencing(t *testing.T) {

	sequencer := NewSequencer(10)
	sequencer.Start()

	var wg sync.WaitGroup
	wg.Add(10)

	requests := make([]*http.InsertRequest, 10)
	for i := 0; i < 10; i++ {
		data := http.InsertData{Message: fmt.Sprintf("Event %d", i)}
		requests[i] = &http.InsertRequest{
			InsertData:      data,
			ResponseChannel: make(chan *http.InsertResponse),
		}
	}

	for i, req := range requests {
		go func(index int, request *http.InsertRequest) {
			response := <-request.ResponseChannel
			// FIXME:+1 is due that the three is a singleton and must be encapsulated
			if response.Index != uint64(index)+1 {
				t.Errorf("The assigned index doesn't obey the insertion order %d, %d", response.Index, index)
			}
			wg.Done()
		}(i, req)
		sequencer.Enqueue(req)
	}

	wg.Wait()
	sequencer.Stop()

}

func BenchmarkSingleInsertion(b *testing.B) {

	data := http.InsertData{Message: "Event 1"}
	request := &http.InsertRequest{
		InsertData:      data,
		ResponseChannel: make(chan *http.InsertResponse),
	}

	sequencer := NewSequencer(10)
	sequencer.Start()

	for i := 0; i < b.N; i++ {
		sequencer.Enqueue(request)
		<-request.ResponseChannel
	}

	sequencer.Stop()
}
