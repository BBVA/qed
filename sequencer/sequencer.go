package sequencer

import (
	"encoding/hex"

	"github.com/golang/glog"

	"verifiabledata/api/http"
	"verifiabledata/merkle/history"
)

type Sequencer struct {
	Tree               *history.Tree
	InsertRequestQueue chan *http.InsertRequest
	FetchRequestQueue  chan *http.FetchRequest
	QuitChan           chan bool
}

func NewSequencer(bufferSize uint, tree *history.Tree) Sequencer {
	sequencer := Sequencer{
		Tree:               tree,
		InsertRequestQueue: make(chan *http.InsertRequest, bufferSize),
		FetchRequestQueue:  make(chan *http.FetchRequest, bufferSize),
		QuitChan:           make(chan bool),
	}
	return sequencer
}

func (sequencer *Sequencer) Start() {

	// FIXME: temporal mock insead of the SMT
	var smt_mock = make(map[string]*http.InsertResponse)

	go func() {
		for {
			select {
			case request := <-sequencer.InsertRequestQueue:
				// glog.Infof("Handling event: %s", request.InsertData.Message)

				commitment, node, err := sequencer.Tree.Add([]byte(request.InsertData.Message))
				if err != nil {
					panic(err)
				}

				response := http.InsertResponse{
					Hash:       string(node.Digest),
					Commitment: string(commitment.Digest),
					Index:      node.Pos.Index,
				}

				// FIXME: temporal mock insead of the SMT
				smt_mock[request.InsertData.Message] = &response

				glog.Infof("New event inserted with index [%d]: %s", response.Index,
					hex.EncodeToString([]byte(response.Commitment)))

				request.ResponseChannel <- &response

			case fetch := <-sequencer.FetchRequestQueue:
				glog.Infof("Fetching  event: %s", fetch.FetchData.Message)
				// FIXME: temporal mock insead of the SMT
				fetch.ResponseChannel <- &http.FetchResponse{
					Index: smt_mock[fetch.FetchData.Message].Index,
				}

			case <-sequencer.QuitChan:
				return
			}
		}
	}()
}

func (sequencer *Sequencer) Stop() {
	glog.Infof("Stopping sequencer...")
	go func() {
		sequencer.QuitChan <- true
	}()
}

func (sequencer *Sequencer) Enqueue(request *http.InsertRequest) {
	glog.Infof("Enqueuing request: %s", request.InsertData.Message)
	sequencer.InsertRequestQueue <- request
}
