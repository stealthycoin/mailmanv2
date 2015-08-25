package mailmanv2

import (
	"fmt"
	"log"
	"sync"
	"net/http"
	"net/url"
	"encoding/json"
)

// CollectRequest can be used to queue a work request without going through
// the mailman server.
var (
	endpoints map[string]endpoint
	methods map[string]pushmethod
	requests map[string]*WorkRequest
	workerQueue chan chan *WorkRequest
	workQueue chan *WorkRequest
	workers []*Worker
	collectorQuit chan bool
	CollectRequest chan *WorkRequest
	backup chan bool
	wg sync.WaitGroup
)


//
// Add an endpoint function
//
func AddEndpoint(name string, f endpoint) {
	endpoints[name] = f
}


//
// Add a method type
//
func AddMethod(name string, f pushmethod) {
	methods[name] = f
}


//
//Create and launch a collector
//
func InitCollector(workerCount int) {
	// Init fn maps
	endpoints = make(map[string]endpoint)
	methods = make(map[string]pushmethod)

	// Global values
	requests    = make(map[string]*WorkRequest)
	workerQueue = make(chan chan *WorkRequest, workerCount)
	workQueue   = make(chan *WorkRequest, 100)
	workers     = make([]*Worker, 0, 0)

	collectorQuit = make(chan bool)
	CollectRequest = make(chan *WorkRequest, 256)
	backup = make(chan bool)

	wg.Add(workerCount)

	for i := 1 ; i <= workerCount ; i++ {
		w := NewWorker(i, workerQueue)
		w.Start()
		workers = append(workers, w)
	}

	// Wait for work requests (this happens after a work request timer has expired)
	go func() {
		for {
			select {
			case work := <- workQueue:
				// Make sure work is still valid
				if work == requests[work.Uid] {
					delete(requests, work.Uid)
					go func() {
						// Get a worker from the worker queue
						worker := <- workerQueue

						// Give the worker the work to do
						worker <- work
					}()
				}
			case <- collectorQuit:
				for _, worker := range workers {
					worker.Quit <- true
				}
				return
			case wr := <- CollectRequest:
				// Check method
				if wr.Method == "add" {
					// Add a new work request, replacing any old one with the same uid
					if oldwr, ok := requests[wr.Uid]; ok {
						oldwr.Cancel <- true
					}
					requests[wr.Uid] = wr
					go wr.StartTimer()
				} else if wr.Method == "cancel" {
					// Remove a work request with a given uid if it exists
					if oldwr, ok := requests[wr.Uid]; ok {
						oldwr.Cancel <- true
						delete(requests, wr.Uid)
					}
				} else {
					if fn, ok := methods[wr.Method]; ok {
						old, ok := requests[wr.Uid]
						wr = fn(old, wr)
						if ok {
							old.Cancel <- true
						}
						requests[wr.Uid] = wr
						go wr.StartTimer()
					} else {
						log.Println("No such method:", wr.Method)
					}
				}
			case <- backup:
				// do some backup stuff
				BackupRequests()
			}
		}
	}()
}

// Handle a work request being sent
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	// Recover from errors
	defer func() {
		if rec := recover() ; rec != nil {
			log.Println(rec)
		}
	}()

	var wr WorkRequest
	err := json.Unmarshal([]byte(r.FormValue("work")), &wr)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
	} else {
		wr.Cancel = make(chan bool)
		CollectRequest <- &wr
		w.WriteHeader(200)
	}

}

func Remove(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(r.URL.String())
	if u.Query().Get("password") == "pancakesauce" {
		hash := u.Query().Get("hash")
		if r, ok := requests[hash] ; ok {
			r.Cancel <- true
			delete(requests, hash)
		}
	}
	fmt.Fprintf(w, "<script>window.close()</script>")
}

// Show me the goods
func ShowMeTheGoods(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(r.URL.String())

	if u.Query().Get("password") == "pancakesauce" {
		if len(requests) == 0 {
			fmt.Fprintf(w, "No mail")
		}
		for _, mail := range requests {
			fmt.Fprintf(w, "<div onclick='window.open(\"/remove?password=pancakesauce&hash=%v\");location.reload();' style='margin-right:2px;text-align:center;display:inline-block;width:1em;border:1px solid black;border-radius:50%%'>x</div>%v %v %v %v<br/>", mail.Uid, mail.Endpoint, mail.Token, mail.Timestamp, mail.Payload)
		}
	} else {
		fmt.Fprintf(w, "404 page not found")
	}
}

// Shutdown collector
func StopCollector() {
	collectorQuit <- true
	wg.Wait()
	fmt.Println("All workers shut down.")
}
