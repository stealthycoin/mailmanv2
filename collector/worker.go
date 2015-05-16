package collector

import (
	"fmt"
	"log"
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
	// Config worker
	w := &Worker{
		Id: id,
		Work: make(chan *WorkRequest),
		WorkerQueue: workerQueue,
		Quit: make(chan bool),
	}

	w.OpenAPNS()

	return w
}

func (w *Worker) OpenAPNS() {
	// load test cert/key
	testCertPem, err := ioutil.ReadFile(Config["apple_push_test_cert"])
	if err != nil {
		log.Fatal(err)
	}
	testKeyPem, err := ioutil.ReadFile(Config["apple_push_test_key"])
	if err != nil {
		log.Fatal(err)
	}
	tc, err := apns.NewAPNSConnection(&apns.APNSConfig{
		CertificateBytes: testCertPem,
		KeyBytes: testKeyPem,
		GatewayHost: "gateway.sandbox.push.apple.com",
	})
	if err != nil {
		log.Fatal(err)
	}



	// load cert/key
	certPem, err := ioutil.ReadFile(Config["apple_push_cert"])
	if err != nil {
		log.Fatal(err)
	}
	keyPem, err := ioutil.ReadFile(Config["apple_push_key"])
	if err != nil {
		log.Fatal(err)
	}
	rc, err := apns.NewAPNSConnection(&apns.APNSConfig{
		CertificateBytes: certPem,
		KeyBytes: keyPem,
	})
	if err != nil {
		log.Fatal(err)
	}

	w.apns_test = tc
	w.apns_real = rc
}

func (w *Worker) Start() {
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
				w.apns_test.Disconnect()
				w.apns_real.Disconnect()
				wg.Done()
				return
			}
		}
	}()
}
