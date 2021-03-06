package mailmanv2

import (
	"log"
	"time"
	apns "github.com/joekarl/go-libapns"
)

type WorkRequest struct {
	Uid string       `json:"uid"`
	Endpoint string  `json:"endpoint"`
	Method string    `json:"method"`
	Payload string   `json:"payload"`
	Token string     `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Cancel chan bool `json:"-"`
	apns_test *apns.APNSConnection `json:"-"`
	apns_real *apns.APNSConnection `json:"-"`
}


func NewWorkRequest(uid, endpoint, method, token, payload string, timestamp int64) *WorkRequest {
	return &WorkRequest{
		Uid:       uid,
		Endpoint:  endpoint,
		Method:    method,
		Payload:   payload,
		Token:     token,
		Timestamp: timestamp,
		Cancel:    make(chan bool),
	}
}



func (wr *WorkRequest) TryCancel() {
	defer func() {
		if r := recover() ; r != nil {
			log.Println("Cannot cancel", wr)
		}
	}()
	wr.Cancel <- true
}



func (wr *WorkRequest) StartTimer() {
	sleepTime := wr.Timestamp - time.Now().Unix()
	if sleepTime <= 0 {
		log.Println("Queueing right away", wr)
		requests[wr.Uid] = wr
		workQueue <- wr
	} else {
		// Wait for a cancel or for the timer to expire
		requests[wr.Uid] = wr
		go func() {
			timer := time.NewTimer(time.Duration(sleepTime) * time.Second)
			select {
			case <- wr.Cancel:
				timer.Stop()
			case <- timer.C:
				log.Println("Queueing after sleep", wr)
				close(wr.Cancel)
				workQueue <- wr
			}
		}()
	}
}
