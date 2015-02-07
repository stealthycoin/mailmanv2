package main

var (
	requests map[string]*WorkRequest
	workerQueue chan chan *WorkRequest
	workQueue chan *WorkRequest
	workers []*Worker
)


type Payload struct {
	Message string `json:"message"`
}

// Create and launch a collector
func InitCollector(workerCount int) {
	requests = make(map[string]*WorkRequest)
	workerQueue = make(chan chan *WorkRequest, workerCount)
	workQueue = make(chan *WorkRequest, 100)
	workers = make([]*Worker, 0, 0)

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
				go func() {
					// Get a worker from the worker queue
					worker := <- workerQueue

					// Give the worker the work to do
					worker <- work
				}()
			}
		}
	}()
}

// Accept a request for work
func IssueWorkRequest(r *WorkRequest) {
	// Check for existing routine with same uid
	// if it exists tell it to cancel
	if wr, ok := requests[r.uid] ; ok {
		wr.cancel <- true
	}
	go r.StartTimer()
}
