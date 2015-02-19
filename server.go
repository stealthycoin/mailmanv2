package main

import (
	"log"
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
			log.Println(wr)
			collectRequest <- &wr
			w.WriteHeader(200)
		}
	})

	http.ListenAndServe(":8003", nil)
}
