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
	wc, _ := strconv.Atoi(config["workers"])
	InitCollector(wc)

	// Handler function for requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var wr WorkRequest
		err := json.Unmarshal([]byte(r.FormValue("work")), &wr)
		if err != nil {
			collectRequest <- &wr
		}
	})

	http.ListenAndServe(":8003", nil)
}
