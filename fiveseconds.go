package main

import (
	"time"
	"strconv"
)

func main() {
	InitConfig()
	InitCollector(1)
	InitPersist()
	fiveSeconds := time.Now().UnixNano() + 5000000000
	collectRequest <- NewWorkRequest(strconv.FormatInt(time.Now().UnixNano(), 10), "println", "3", "This is five seconds later", fiveSeconds)
	time.Sleep(time.Duration(6) * time.Second)
	StopCollector()
}
