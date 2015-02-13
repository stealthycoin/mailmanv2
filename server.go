package main

import (
	"strconv"
	"net/http"
	"encoding/json"
)

func main() {
	// Init all the components
	InitConfig()
	InitPersist()
	wc, _ := strconv.Atoi(config["workers"].(string))
	InitCollector(wc)

	// Handler function for requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var wr WorkRequest
		err := json.Unmarshal([]byte(r.FormValue("work")), &wr)
		wr.Cancel = make(chan bool)
		wr.Valid = true
		if err != nil {
			IssueWorkRequest(&wr)
		}
	})

	http.ListenAndServe(":8003", nil)
}
