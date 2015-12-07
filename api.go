package mailmanv2

import (
	"fmt"
	"log"
	"strconv"
	"net/http"
	"encoding/json"
)


//
// Func get a list of keys from the active work requests
//
func MMMail(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(requests)
	fmt.Fprintf(w, string(b))
}


//
// Get status of all workers
//
func MMStatus(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(workers)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	fmt.Fprintf(w, string(b))
}


//
// Manually reboot a worker
//
func MMRebootWorker(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	// Find that id since they may not be in order
	var location int = 0
	for index, worker := range workers {
		if worker.Id == id {
			location = index
		}
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Not a valid worker id", 500)
	} else {
		// Tell worker to quit and wait for it to be done
		workers[location].Quit <- true
		<- workers[location].Done

		// Reconstruct a new worker and start it.
		workers[location] = NewWorker(int(id), workerQueue)
		workers[location].Start()

		// They old worker work channel will still be enqueued
		// it will be thrown out by the collector next time it is encountered
		fmt.Fprintf(w, `"Worker rebooted"`)
	}
}
