package main

import (
	"log"
	"strconv"
	"net/http"
	"encoding/json"
	_ "github.com/lib/pq"
	"database/sql"
	"bitbucket.org/mailman/collector"
)

var (
	db *sql.DB
)

func main() {
	// Connect to the databse
	tdb, err := sql.Open("postgres", "dbname=hearth user=hearth host=54.67.5.205 password=A938CEA3C22F8FD93F4157D4A1AB3AF753452D743FEC6A8B27401972B3F9511F sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db = tdb
	collector.SetDb(db)


	// Init all the components
	collector.InitConfig()
	wc, _ := strconv.Atoi(collector.Config["workers"])
	collector.InitCollector(wc)
	collector.InitPersist()

	// Handler function for requests
	http.HandleFunc("/push/", collector.RequestHandler)

	http.ListenAndServe(":8003", nil)
}
