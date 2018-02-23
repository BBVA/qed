package sequencer

import (
	"log"
	"verifiabledata/api"
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
	}
	return sequencer
}

func (sequencer *Sequencer) Start() {
	go func() {
		select {
		case request := <-sequencer.InsertRequestQueue:
			if sequencer.Counter%1000 == 0 {
				log.Printf("Handling event: %s", request.InsertData.Message)
			}
			commitment := util.Hash([]byte(request.InsertData.Message)) // TODO USE BYTE ARRAYS OR STRINGS
			response := &api.InsertResponse{
				Hash:       string(commitment),
				Commitment: string(commitment),
				Index:      sequencer.Counter,
			}
			if sequencer.Counter%1000 == 0 {
				log.Printf("New event inserted with index [%d]: %s", response.Index, response.Commitment)
			}
			sequencer.Counter++
			request.ResponseChannel <- response
		case <-sequencer.QuitChan:
			return
		}
	}()
}

func (sequencer *Sequencer) Stop() {
	go func() {
		sequencer.QuitChan <- true
	}()
}
