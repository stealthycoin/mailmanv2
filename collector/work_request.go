package collector

import (
	"time"
)

type WorkRequest struct {
	Uid string       `json:"uid"`
	Endpoint string  `json:"endpoint"`
	Payload string   `json:"payload"`
	Token string     `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Cancel chan bool
}


func NewWorkRequest(uid, endpoint, token, payload string, timestamp int64) *WorkRequest {
	return &WorkRequest{
		Uid: uid,
		Endpoint: endpoint,
		Payload: payload,
		Token: token,
		Timestamp: timestamp,
		Cancel: make(chan bool),
	}
}

func (wr *WorkRequest) StartTimer() {
	sleepTime := wr.Timestamp - time.Now().Unix()
	timer := time.NewTimer(time.Duration(sleepTime) * time.Second)
	if sleepTime <= 0 {
		workQueue <- wr
	} else {
		// Wait for a cancel or for the timer to expire
		select {
		case <- wr.Cancel:
			timer.Stop()
		case <- timer.C:
			workQueue <- wr
		}
	}
}
