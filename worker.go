package mailmanv2

import (
	"fmt"
	"log"
	"io/ioutil"
	apns "github.com/joekarl/go-libapns"
)

const (
	MAX_BUFFERED_MESSAGES = 100
)

var (
	error_handlers = make(map[string]errorhandler)
)


//
// Buffer to hold sent payloads for error handling
//
type PayloadBuffer struct {
	buffer        []*apns.Payload
	buffer_offset uint32
	error         bool
}


//
// Worker structure
//
type Worker struct {
	Id             int
	Work           chan *WorkRequest
	WorkerQueue    chan chan *WorkRequest
	Quit           chan bool
	// Parallel maps, not good (TODO)
	Apns_cons      map[string]*apns.APNSConnection
	payload_buffer map[string]*PayloadBuffer
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
		Id:             id,
		Work:           make(chan *WorkRequest),
		WorkerQueue:    workerQueue,
		Quit:           make(chan bool),
		Apns_cons:      make(map[string]*apns.APNSConnection),
		payload_buffer: make(map[string]*PayloadBuffer),
	}

	return w
}


//
// Map a pair of config values to apns key/cert pair
/// TODO explicar
//
func NewApns(key string) {
	// For all workers currently running
	for _, w := range workers {
		// Set a blank connection and set its error to true
		w.Apns_cons[key] = nil
		w.payload_buffer[key] = &PayloadBuffer{
			buffer:        make([]*apns.Payload, 0),
			buffer_offset: 1,
			error:         true,
		}
	}
}


//
// Load apns settings
// TODO MAKE USEFUL DOCS HERE MAN
//
func (w *Worker) OpenAPNS(key string) {
	_, ok1 := Config[key + "_key"]
	_, ok2 := Config[key + "_cert"]

	if ok1 && ok2 {
		log.Printf("Opening APNS connection %s for worker %d\n", key, w.Id)

		// load cert/key
		certPem, err := ioutil.ReadFile(Config[key + "_cert"])
		if err != nil {
			log.Fatal(err)
			return
		}
		keyPem, err := ioutil.ReadFile(Config[key + "_key"])
		if err != nil {
			log.Fatal(err)
			return
		}
		gateway := "gateway.push.apple.com"
		if ngate, ok := Config[key + "_gate"]; ok {
			gateway = ngate
		}
		conn, err := apns.NewAPNSConnection(&apns.APNSConfig{
			CertificateBytes: certPem,
			KeyBytes:         keyPem,
			GatewayHost:      gateway,
		})
		if err != nil {
			log.Fatal(err)
			return
		}

		// Add connection to connection map
		w.Apns_cons[key] = conn

		// Create buffer
		w.payload_buffer[key] = &PayloadBuffer{
			buffer:        make([]*apns.Payload, 0),
			buffer_offset: 1,
			error:         false,
		}

		go w.ErrorListen(key)
	} else {
		log.Printf("No such key %s found in config\n", key)
	}
}


//
// Listen for apns errors on a specific connection
//
func (w *Worker) ErrorListen(key string) {
	if _, ok := w.payload_buffer[key]; !ok {
		log.Printf("No such key %s\n", key)
		return
	}

	// If we make it to the end of the function there was an error
	defer func() {
		w.payload_buffer[key].error = true
	}()


	// Fetch close channel for the connection
	cc, ok := <- w.Apns_cons[key].CloseChannel
	if !ok || cc == nil {
		return
	}

	// Handle an error
	handle := func(code string) {
		if eh, ok := error_handlers[code]; ok {
			log.Println("Handling", code)
			pb := w.payload_buffer[key]
			if idx := cc.Error.MessageID - pb.buffer_offset; idx >= 0 && idx < uint32(len(pb.buffer)) {
				eh(pb.buffer[idx])
			} else {
				log.Println("MessageID out of bounds", idx, len(pb.buffer))
			}
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
		handle("INVALID_TOKEN")
	}
}

//
// Place last sent payload into a buffer to reference in case of error
//
func (w *Worker) bufferPayload(key string, payload *apns.Payload) {
	if pb, ok := w.payload_buffer[key]; ok {
		pb.buffer = append(pb.buffer, payload)
		if len(pb.buffer) > MAX_BUFFERED_MESSAGES {
			pb.buffer = pb.buffer[1:]
			pb.buffer_offset++
		}
	}
}


//
// Send over the channel opening if required
//
func (w *Worker) Send(key string, payload *apns.Payload) {
	if pb, ok := w.payload_buffer[key]; ok {
		// Check for error and reopen if required
		if pb.error {
			pb.error = false
			w.OpenAPNS(key)
		}
		// Send message and buffer it
		w.Apns_cons[key].SendChannel <- payload
		w.bufferPayload(key, payload)
	} else {
		log.Printf("Cannot send to channel located at key %s\n", key)
	}
}

//
// Start worker routine
//
func (w *Worker) Start() {
	go func() {
		for {
			log.Println("Queueing worker", w.Id)
			w.WorkerQueue <- w.Work
			select {
			case wr := <- w.Work:
				func () {
					// Recover from any error that occurs during work, must requeue the worker
					defer func() {
						if r := recover(); r != nil {
							log.Println("Worker", w.Id, "recovered", r)
						}
					}()

					// Check for endpoint and call function
					if fn, ok := endpoints[wr.Endpoint]; ok {
						log.Println("Worker", w.Id, "Starting work")
						fn(wr, w)
					} else {
						log.Println("No such endpoint", wr.Endpoint)
					}
				}()
			case <- w.Quit:
				fmt.Printf("Worker %d shutting down.\n", w.Id)
				wg.Done()
				for _, con := range w.Apns_cons {
					con.Disconnect()
				}

				return
			}
		}
	}()
}
