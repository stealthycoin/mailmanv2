package mailmanv2

import (
	"log"
	"flag"
	"strconv"
	"net/http"
)

func StartServer() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	// Init all the components
	InitConfig()
	wc, _ := strconv.Atoi(Config["workers"])
	InitCollector(wc)
	InitPersist()

	// Handler function for requests
	http.HandleFunc("/push/", RequestHandler)

	// Handlers
	http.HandleFunc("/remove", Remove)
	http.HandleFunc("/showmethegoods", ShowMeTheGoods)

	http.ListenAndServe(":8003", nil)
}
