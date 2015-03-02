package collector

import (
	"fmt"
	"sync"
)

// Package Variables
var (
	endpoints map[string]endpoint
	requests map[string]*WorkRequest
	workerQueue chan chan *WorkRequest
	workQueue chan *WorkRequest
	workers []*Worker
	collectorQuit chan bool
	CollectRequest chan *WorkRequest
	backup chan bool
	wg sync.WaitGroup
)


// Create and launch a collector
func InitCollector(workerCount int) {
	endpoints = make(map[string]endpoint)
	// Test endpoints
	endpoints["testtime"] = TestTimePayload
	endpoints["testpayload"] = TestPayload
	endpoints["println"] = PrintlnEndpoint

	// Real endpoints
	endpoints["website"] = WebsiteEndpoint
	endpoints["phone"] = PhoneEndpoint
	endpoints["apns"] = ApnsEndpoint
	endpoints["cancel"] = CancelEndpoint

	// Global values
	requests = make(map[string]*WorkRequest)
	workerQueue = make(chan chan *WorkRequest, workerCount)
	workQueue = make(chan *WorkRequest, 100)
	workers = make([]*Worker, 0, 0)

	collectorQuit = make(chan bool)
	CollectRequest = make(chan *WorkRequest)
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
				if oldwr, ok := requests[wr.Uid] ; ok {
					oldwr.Cancel <- true
				}
				requests[wr.Uid] = wr
				go wr.StartTimer()
			case <- backup:
				// do some backup stuff
				fmt.Println("backup")
				BackupRequests()
			}
		}
	}()
}

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
		collector.collectRequest <- &wr
		w.WriteHeader(200)
	}

}

// Shutdown collector
func StopCollector() {
	collectorQuit <- true
	wg.Wait()
	fmt.Println("All workers shut down.")
}
