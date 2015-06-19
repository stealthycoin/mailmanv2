package mailmanv2

import (
	"strconv"
	"net/http"
)

func StartServer() {
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
