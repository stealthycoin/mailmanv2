package main

import "time"

func main() {
	InitCollector(1)
	InitPersist()
	fiveSeconds := time.Now().UnixNano() + 5000000000
	IssueWorkRequest(NewWorkRequest("ID", "println", "This is five seconds later", fiveSeconds))
	time.Sleep(time.Duration(6) * time.Second)
	StopCollector()
}
