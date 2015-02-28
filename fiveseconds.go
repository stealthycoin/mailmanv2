package main

import (
	"time"
	"database/sql"
)
var (
	db *sql.DB
)

func main() {
	InitConfig()
	InitCollector(1)
	InitPersist()
	fiveSeconds := time.Now().Unix() + 5
	collectRequest <- NewWorkRequest("test", "println", "3", "This is five seconds later", fiveSeconds)
	collectRequest <- NewWorkRequest("test", "println", "3", "This is five seconds later fo realz", fiveSeconds)
	collectRequest <- NewWorkRequest("hamwallet", "println", "3", "This is five seconds later fo BROAH", fiveSeconds)

	time.Sleep(time.Duration(6) * time.Second)
	StopCollector()
}
