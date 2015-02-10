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
		Uid: uid,
		Endpoint: endpoint,
		Payload: payload,
		Timestamp: timestamp,
		Cancel: make(chan bool),
		Valid: true,

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
				if wr.Valid {
					workQueue <- wr
				}
				running = false
			case <- wr.Cancel:
				wr.Valid = false
			}
		}
	}()

	// Trigger the timeout in the listening routine after sleeping
	sleepTime := wr.Timestamp - time.Now().UnixNano()
	if sleepTime <= 0 {
		timeout <- true
	} else {
		time.Sleep(time.Duration(sleepTime) * time.Nanosecond)
		timeout <- true
	}
}
