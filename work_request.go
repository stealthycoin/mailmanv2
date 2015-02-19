package main

import (
	"time"
)

type WorkRequest struct {
	Uid string       `json:"uid"`
	Endpoint string  `json:"endpoint"`
	Payload string   `json:"payload"`
	Token string     `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Testing bool     `json:"testing"` // Just for apple push notes brah
}


// Construct a new work request (this is for testing purposes)
func NewWorkRequest(uid, endpoint, token, payload string, timestamp int64) *WorkRequest {
	return &WorkRequest{
		Uid: uid,
		Endpoint: endpoint,
		Payload: payload,
		Token: token,
		Timestamp: timestamp,
		Testing: true,
	}
}


// Waiting process
func (wr *WorkRequest) StartTimer() {
	timeout := make(chan bool)

	// Start listening routine
	go func() {
		<- timeout
		workQueue <- wr
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
