package main

import (
	"time"
)

type WorkRequest struct {
	uid string
	payload string
	timeout int
	start_time int64
	cancel chan bool
	valid bool
}


// Construct a new work request
func NewWorkRequest(uid, payload string, timeout int) *WorkRequest {
	return &WorkRequest{
		uid: uid,
		payload: payload,
		timeout: timeout,
		start_time: time.Now().UnixNano(),
		cancel: make(chan bool),
		valid: true,

	}
}


// Waiting process
func (wr *WorkRequest) StartTimer() {
	timeout := make(chan bool)
	go func() {
		time.Sleep(time.Duration(wr.timeout) * time.Second)
		timeout <- true
	}()
	running := true
	for running {
		select {
		case <- timeout:
			if wr. valid {
				workQueue <- wr
			}
			running = false
		case <- wr.cancel:
			wr.valid = false
		}
	}
}
