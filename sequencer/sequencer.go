package sequencer

type Processer interface {
	Process()
}

type Sequencer struct {
	SyncRequest  chan Processer
	AsyncRequest chan Processer
	QuitChan     chan bool
}

func NewSequencer(bufferSize uint) Sequencer {
	sequencer := Sequencer{
		SyncRequest:  make(chan Processer, bufferSize),
		AsyncRequest: make(chan Processer, bufferSize),
		QuitChan:     make(chan bool),
	}
	return sequencer
}

func (sequencer *Sequencer) Start() {
	go func() {
		for {
			select {
			case request := <-sequencer.SyncRequest:
				request.Process()

			case request := <-sequencer.AsyncRequest:
				go request.Process()

			case <-sequencer.QuitChan:
				return
			}
		}
	}()
}

func (sequencer *Sequencer) Stop() {
	go func() {
		sequencer.QuitChan <- true
	}()
}

func (sequencer *Sequencer) Enqueue(request Processer) {
	sequencer.SyncRequest <- request
}
