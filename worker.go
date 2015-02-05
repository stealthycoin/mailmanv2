package main

import (
	"fmt"
)

type Worker struct {
	Id int
	Work chan *WorkRequest
	WorkerQueue chan chan *WorkRequest
	Quit chan bool
}

func NewWorker(id int, workerQueue chan chan *WorkRequest) *Worker {
	return &Worker{
		Id: id,
		Work: make(chan *WorkRequest),
		WorkerQueue: workerQueue,
		Quit: make(chan bool),
	}
}


func (w* Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case wr := <- w.Work:
				TestPayload(wr)
			case <- w.Quit:
				fmt.Printf("Worker %d shutting down.\n", w.Id)
				return
			}
		}
	}()
}
