package sequencer

import (
	"encoding/hex"
	"log"

	"verifiabledata/api/http"
	"verifiabledata/merkle/history"
)

type Sequencer struct {
	Tree               *history.Tree
	InsertRequestQueue chan *http.InsertRequest
	QuitChan           chan bool
}

func NewSequencer(bufferSize uint, tree *history.Tree) Sequencer {
	sequencer := Sequencer{
		Tree:               tree,
		InsertRequestQueue: make(chan *http.InsertRequest, bufferSize),
		QuitChan:           make(chan bool),
	}
	return sequencer
}

func (sequencer *Sequencer) Start() {
	go func() {
		for {
			select {
			case request := <-sequencer.InsertRequestQueue:
				log.Printf("Handling event: %s", request.InsertData.Message)

				commitment, node, err := sequencer.Tree.Add([]byte(request.InsertData.Message))
				if err != nil {
					panic(err)
				}

				response := http.InsertResponse{
					Hash:       string(node.Digest),
					Commitment: string(commitment.Digest),
					Index:      node.Pos.Index,
				}

				log.Printf("New event inserted with index [%d]: %s", response.Index,
					hex.EncodeToString([]byte(response.Commitment)))

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

func (sequencer *Sequencer) Enqueue(request *http.InsertRequest) {
	log.Printf("Enqueuing request: %s", request.InsertData.Message)
	sequencer.InsertRequestQueue <- request
}
