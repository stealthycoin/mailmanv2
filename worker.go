package mailmanv2

import (
	"fmt"
	"log"
	"io/ioutil"
	apns "github.com/joekarl/go-libapns"
)

var (
	error_handlers = make(map[string]errorhandler)
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
	Error bool
	buffer []*apns.Payload
	buffer_offset uint32
}


//
// Add an error handling function
//
func AddErrorHandler(name string, f errorhandler) {
	error_handlers[name] = f
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
		Error: false,
	}

	w.OpenAPNS()

	return w
}


//
// Load apns settings
//
func (w *Worker) OpenAPNS() {
	log.Printf("Opening APNS connection for worker %d\n", w.Id)
	if _, ok := Config["apple_push_test_cert"]; ok {
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

		w.apns_test = tc
	}


	if _, ok := Config["apple_push_cert"]; ok {
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

		w.apns_real = rc

		// Create buffer
		w.buffer = make([]*apns.Payload, 0)
		w.buffer_offset = 1

		go w.ErrorListen()
	}

}


//
// Listen for apns errors and reload it
//
func (w *Worker) ErrorListen() {
	cc := <- w.apns_real.CloseChannel

	// Handle an error
	handle := func(code string) {
		if eh, ok := error_handlers[code]; ok {
			eh(w.buffer[cc.Error.MessageID - w.buffer_offset])
		} else {
			log.Println("No handler for", code)
		}
	}

	// Which error is it
	switch cc.Error.ErrorCode {
	case 251:
		log.Println("EOF")
		handle("EOF")
	case 1:
		log.Println("PROCESSING_ERROR")
		handle("PROCESSING_ERROR")
	case 2:
		log.Println("MISSING_DEVICE_TOKEN")
		handle("MISSING_DEVICE_TOKEN")
	case 3:
		log.Println("MISSING_TOPIC")
		handle("MISSING_TOPIC")
	case 4:
		log.Println("MISSING_PAYLOAD")
		handle("MISSING_PAYLOAD")
	case 5:
		log.Println("INVALID_TOKEN_SIZE")
		handle("INVALID_TOKEN_SIZE")
	case 6:
		log.Println("INVALID_TOPIC_SIZE")
		handle("INVALID_TOKEN_SIZE")
	case 7:
		log.Println("INVALID_PAYLOAD_SIZE")
		handle("INVALID_PAYLOAD_SIZE")
	case 8:
		log.Println("INVALID_TOKEN")
		hanlde("INVALID_TOKEN")
	}

	// Set error flag to true to restart connection
	w.Error = true
}


//
// Add to buffer of messagse
//
func (w *Worker) BufferPayload(payload *apns.Payload) {
	w.buffer = append(w.buffer, payload)
	if len(w.buffer) > 100 {
		w.buffer = w.buffer[1:]
		w.buffer_offset++
	}
}


//
// Send over the channel opening if required
//
func (w *Worker) Send(payload *apns.Payload, testing bool) {
	if w.Error {
		w.Error = false
		w.OpenAPNS()
	}
	if testing {
		w.apns_test.SendChannel <- payload
	}else {
		w.apns_real.SendChannel <- payload
		w.BufferPayload(payload)
	}
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
				} else {
					log.Println("No such endpoint", wr.Endpoint)
				}
			case <- w.Quit:
				fmt.Printf("Worker %d shutting down.\n", w.Id)
				wg.Done()
				if w.apns_test != nil {
					w.apns_test.Disconnect()
				}
				if w.apns_real != nil {
					w.apns_real.Disconnect()
				}

				return
			}
		}
	}()
}
