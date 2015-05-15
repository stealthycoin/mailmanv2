package main

import (
	"log"
	"flag"
	"strconv"
	"net/http"
	_ "github.com/lib/pq"
	"database/sql"
	"bitbucket.org/mailman/collector"
)

var (
	local = flag.Bool("local", false, "Set when you want to connect to a local db")
	db *sql.DB
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()
	var err error = nil
	var tdb *sql.DB
	if *local {
		// COnnect to local db for testing
		tdb, err = sql.Open("postgres", "dbname=hearth user=hearth host=localhost password=A938CEA3C22F8FD93F4157D4A1AB3AF753452D743FEC6A8B27401972B3F9511F sslmode=disable")
	} else {
		// Connect to real database at amazon
		tdb, err = sql.Open("postgres", "dbname=hearth user=hearth host=50.18.206.141 password=A938CEA3C22F8FD93F4157D4A1AB3AF753452D743FEC6A8B27401972B3F9511F sslmode=require")
	}
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
	collector.InitRecords(collector.Config["record_timeout"])

	// Handler function for requests
	http.HandleFunc("/push/", collector.RequestHandler)

	// Handlers
	http.HandleFunc("/remove", collector.Remove)
	http.HandleFunc("/showmethegoods", collector.ShowMeTheGoods)

	http.ListenAndServe(":8003", nil)
}
