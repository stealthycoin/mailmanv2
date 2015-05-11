package collector

import (
	"fmt"
	"io/ioutil"
	apns "github.com/joekarl/go-libapns"
)

type Worker struct {
	Id int
	Work chan *WorkRequest
	WorkerQueue chan chan *WorkRequest
	Quit chan bool
	apns_test, apns_real *apns.APNSConnection
}

func NewWorker(id int, workerQueue chan chan *WorkRequest) *Worker {
	// load test cert/key
	testCertPem, err := ioutil.ReadFile(Config["apple_push_test_cert"])
	if err != nil {
		panic(err)
	}
	testKeyPem, err := ioutil.ReadFile(Config["apple_push_test_cert"])
	if err != nil {
		panic(err)
	}

	// load cert/key
	certPem, err := ioutil.ReadFile(Config["apple_push_cert"])
	if err != nil {
		panic(err)
	}
	keyPem, err := ioutil.ReadFile(Config["apple_push_cert"])
	if err != nil {
		panic(err)
	}

	// Config worker
	return &Worker{
		Id: id,
		Work: make(chan *WorkRequest),
		WorkerQueue: workerQueue,
		Quit: make(chan bool),
		apns_test: apns.NewAPNSConnection(&APNSConfig{
			CertificateBytes: testCertPem,
			KeyBytes: testKeyPem,
			GatewayHost: "gateway.sandbox.push.apple.com:2195",
		}),
		apns_real:apns.NewAPNSConnection(&APNSConfig{
			CertificateBytes: certPem,
			KeyBytes: keyPem,
		}),
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
