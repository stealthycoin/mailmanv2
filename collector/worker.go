package collector

import (
	"fmt"
	"log"
	"io/ioutil"
	apns "github.com/joekarl/go-libapns"
)


//
// Worker structure
//
type Worker struct {
	Id int
	Work chan *WorkRequest
	WorkerQueue chan chan *WorkRequest
	Quit chan bool
	apns_test, apns_real *apns.APNSConnection
}


//
// Create a new worker
//
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


//
// Load apns settings
//
func (w *Worker) OpenAPNS() {
	log.Printf("Opening APNS connection for worker %d\n", w.Id)
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

	go w.ErrorListen()
}


//
// Listen for apns errors and reload it
//
func (w *Worker) ErrorListen() {
	cc := <- w.apns_real.CloseChannel
	log.Println(cc)
	w.OpenAPNS()
}


//
// Start worker routine
//
func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case wr := <- w.Work:
				if fn, ok := endpoints[wr.Endpoint]; ok {
					fn(wr, w)
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
