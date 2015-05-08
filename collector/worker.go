package collector

import (
	"fmt"
	"github.com/anachronistic/apns"
)

type Worker struct {
	Id int
	Work chan *WorkRequest
	WorkerQueue chan chan *WorkRequest
	Quit chan bool
	apns_test, apns_real *apns.Client
}

func NewWorker(id int, workerQueue chan chan *WorkRequest) *Worker {
	return &Worker{
		Id: id,
		Work: make(chan *WorkRequest),
		WorkerQueue: workerQueue,
		Quit: make(chan bool),
		apns_test: apns.NewClient("gateway.sandbox.push.apple.com:2195",
			Config["apple_push_test_cert"],
			Config["apple_push_test_key"]),
		apns_real:apns.NewClient("gateway.push.apple.com:2195",
			Config["apple_push_cert"],
			Config["apple_push_key"]),

	}
}


func (w* Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case wr := <- w.Work:
				if fn, ok := endpoints[wr.Endpoint]; ok {
					wr.apns_test = w.apns_test
					wr.apns_real = w.apns_real
					fn(wr)
				}
			case <- w.Quit:
				fmt.Printf("Worker %d shutting down.\n", w.Id)
				wg.Done()
				return
			}
		}
	}()
}
