package collector

import (
	"time"
	"github.com/anachronistic/apns"
)

type WorkRequest struct {
	Uid string       `json:"uid"`
	Endpoint string  `json:"endpoint"`
	Method string    `json:"method"`
	Payload string   `json:"payload"`
	Token string     `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Cancel chan bool
	apns_test, apns_real *apns.Client
}


func NewWorkRequest(uid, endpoint, method, token, payload string, timestamp int64) *WorkRequest {
	return &WorkRequest{
		Uid: uid,
		Endpoint: endpoint,
		Method: method,
		Payload: payload,
		Token: token,
		Timestamp: timestamp,
		Cancel: make(chan bool),
		apns_test: apns.NewClient("gateway.sandbox.push.apple.com:2195",
			Config["apple_push_test_cert"],
			Config["apple_push_test_key"]),
		apns_real:apns.NewClient("gateway.push.apple.com:2195",
			Config["apple_push_cert"],
			Config["apple_push_key"]),
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
