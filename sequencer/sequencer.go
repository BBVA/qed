package sequencer

import (
	"encoding/hex"
	"log"
	api "verifiabledata/api/http"
	"verifiabledata/util"
)

type Sequencer struct {
	Counter            uint64
	InsertRequestQueue chan *api.InsertRequest
	QuitChan           chan bool
}

func NewSequencer(bufferSize uint) Sequencer {
	sequencer := Sequencer{
		Counter:            0,
		InsertRequestQueue: make(chan *api.InsertRequest, bufferSize),
		QuitChan:           make(chan bool),
	}
	return sequencer
}

func (sequencer *Sequencer) Start() {
	go func() {
		hasher := util.Hash256()
		for {
			select {
			case request := <-sequencer.InsertRequestQueue:
				//if sequencer.Counter%1000 == 0 {
				log.Printf("Handling event: %s", request.InsertData.Message)
				//}
				commitment := hasher.Do([]byte(request.InsertData.Message)) // TODO USE BYTE ARRAYS OR STRINGS
				response := api.InsertResponse{
					Hash:       string(commitment),
					Commitment: string(commitment),
					Index:      sequencer.Counter,
				}
				//if sequencer.Counter%1000 == 0 {
				log.Printf("New event inserted with index [%d]: %s", response.Index,
					hex.EncodeToString([]byte(response.Commitment)))
				//}
				sequencer.Counter++
				request.ResponseChannel <- &response
			case <-sequencer.QuitChan:
				return
			}
		}
	}()
}

func (sequencer *Sequencer) Stop() {
	log.Printf("Stopping sequencer...")
	go func() {
		sequencer.QuitChan <- true
	}()
}

func (sequencer *Sequencer) Enqueue(request *api.InsertRequest) {
	log.Printf("Enqueuing request: %s", request.InsertData.Message)
	sequencer.InsertRequestQueue <- request
}
