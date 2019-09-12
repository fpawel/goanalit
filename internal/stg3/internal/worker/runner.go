package worker

import "github.com/fpawel/procmq"

type Runner struct {
	peer              *procmq.ProcessMQ
	portName          chan string
	placeAddr         chan placeAddr
	setDeviceVarValue chan setDeviceVarValue
}

func NewRunner(peer *procmq.ProcessMQ) Runner {
	return Runner{
		peer:              peer,
		placeAddr:         make(chan placeAddr),
		setDeviceVarValue: make(chan setDeviceVarValue),
		portName:          make(chan string),
	}
}

func (x Runner) Run() {
	w := newWorker(x.peer)

	for {
		if err := w.openPort(); err != nil {
			x.peer.Send("ERROR", err.Error())
		}
	}

}
