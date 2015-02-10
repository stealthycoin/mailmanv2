package main

import (
	"time"
_	"fmt"
)

type WorkRequest struct {
	Uid string
	Endpoint string
	Payload string
	Timestamp int64
	Cancel chan bool
	Valid bool
}


// Construct a new work request
func NewWorkRequest(uid, endpoint, payload string, timestamp int64) *WorkRequest {
	return &WorkRequest{
		uid: uid,
		endpoint: endpoint,
		payload: payload,
		timestamp: timestamp,
		cancel: make(chan bool),
		valid: true,

	}
}


// Waiting process
func (wr *WorkRequest) StartTimer() {
	timeout := make(chan bool)

	// Start listening routine
	go func() {
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
	}()

	// Trigger the timeout in the listening routine after sleeping
	sleepTime := wr.timestamp - time.Now().UnixNano()
	if sleepTime <= 0 {
		timeout <- true
	} else {
		time.Sleep(time.Duration(sleepTime) * time.Nanosecond)
		timeout <- true
	}
}
