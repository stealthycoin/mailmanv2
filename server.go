package mailmanv2

import (
	"strconv"
	"net/http"
)


//
// Init all mailman modules
//
func Init() {
	InitConfig()
	wc, _ := strconv.Atoi(Config["workers"])
	InitCollector(wc)
	InitPersist()
}


//
// Registers routes and launches the server
//
func StartServer() {
	// Handler function for requests
	http.HandleFunc("/push/", RequestHandler)

	// Handlers
	http.HandleFunc("/remove", Remove)
	http.HandleFunc("/showmethegoods", ShowMeTheGoods)


	// API
	http.HandleFunc("/mm/api/mail", MMMail)
	http.HandleFunc("/mm/api/status", MMStatus)
	http.HandleFunc("/mm/api/reboot_worker", MMRebootWorker)

	http.ListenAndServe(":8003", nil)
}
