package main

import (
	"log"
	"flag"
	"strconv"
	"net/http"
	"bitbucket.org/mailmanv2/collector"
)

var (
	local = flag.Bool("local", false, "Set when you want to connect to a local db")
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	// Init all the components
	collector.InitConfig()
	wc, _ := strconv.Atoi(collector.Config["workers"])
	collector.InitCollector(wc)
	collector.InitPersist()

	// Handler function for requests
	http.HandleFunc("/push/", collector.RequestHandler)

	// Handlers
	http.HandleFunc("/remove", collector.Remove)
	http.HandleFunc("/showmethegoods", collector.ShowMeTheGoods)

	http.ListenAndServe(":8003", nil)
}
