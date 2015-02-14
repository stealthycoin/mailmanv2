package main

var (
	endpoints map[string]endpoint
	requests map[string]*WorkRequest
	workerQueue chan chan *WorkRequest
	workQueue chan *WorkRequest
	workers []*Worker
	collectorQuit chan bool
	collectRequest chan *WorkRequest
	backup chan bool
)


// Create and launch a collector
func InitCollector(workerCount int) {

	endpoints = make(map[string]endpoint)
	endpoints["testtime"] = TestTimePayload
	endpoints["testpayload"] = TestPayload
	endpoints["println"] = PrintlnEndpoint


	// Global values
	requests = make(map[string]*WorkRequest)
	workerQueue = make(chan chan *WorkRequest, workerCount)
	workQueue = make(chan *WorkRequest, 100)
	workers = make([]*Worker, 0, 0)

	collectorQuit = make(chan bool)
	collectRequest = make(chan *WorkRequest)
	backup = make(chan bool)

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
			case wr := <- collectRequest:
				requests[wr.Uid] = wr
				go wr.StartTimer()
			case <- backup:
				// do some backup stuff

			}
		}
	}()
}

// Shutdown collector
func StopCollector() {
	collectorQuit <- true
}
