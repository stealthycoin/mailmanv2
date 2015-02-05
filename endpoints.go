package main

import (
	"fmt"
	"time"
)

func PrintPayload(wr * WorkRequest) {
	wg.Done()
	fmt.Print(wr.uid + ": took ")
	fmt.Print((time.Now().UnixNano() - wr.start_time) / int64(time.Second), "s ")
	fmt.Println(wr.payload)
}
