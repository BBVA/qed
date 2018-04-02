package sequencer

import (
	"github.com/golang/glog"

	"verifiabledata/api/http"
)

type Sequencer struct {
	InsertRequestQueue chan *http.InsertRequest
	FetchRequestQueue  chan *http.FetchRequest
	QuitChan           chan bool
}

func NewSequencer(bufferSize uint) Sequencer {
	sequencer := Sequencer{
		InsertRequestQueue: make(chan *http.InsertRequest, bufferSize),
		FetchRequestQueue:  make(chan *http.FetchRequest, bufferSize),
		QuitChan:           make(chan bool),
	}
	return sequencer
}

func (sequencer *Sequencer) Start() {

	go func() {
		for {
			select {
			case request := <-sequencer.InsertRequestQueue:
				glog.Infof("Handling event: %s", request.InsertData.Message)
				request.ProcessResponse(request.InsertData, request.ResponseChannel)

			case fetch := <-sequencer.FetchRequestQueue:
				glog.Infof("Fetching  event: %s", fetch.FetchData.Message)
				go fetch.ProcessResponse(fetch.FetchData, fetch.ResponseChannel)

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
