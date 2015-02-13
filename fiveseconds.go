package main

import (
	"time"
	"fmt"
//	"strconv"
)

func main() {
	InitCollector(1)
	InitPersist()
	fiveSeconds := time.Now().UnixNano() + 5000000000
	fmt.Println(requests)
	IssueWorkRequest(NewWorkRequest(strconv.FormatInt(time.Now().UnixNano(), 10), "println", "This is five seconds later", fiveSeconds))
	collectRequest <- NewWorkRequest("ID", "println", "This is five seconds later", fiveSeconds)
	fmt.Println(requests)
	time.Sleep(time.Duration(6) * time.Second)
	StopCollector()
}
