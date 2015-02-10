package main

import (
	"fmt"
	"time"
	"bytes"
	"strconv"
)

var (
	TestResults chan string
)

type endpoint func(wr *WorkRequest)

func TestTimePayload(wr *WorkRequest) {
	var buffer bytes.Buffer

	// Write testing message
	diff, err := strconv.ParseInt(wr.Payload, 10, 64)
	if err != nil {
		TestResults <- "Oh fuck"
	} else {
		diff -= time.Now().UnixNano()
		diff /= 1000000000

		buffer.WriteString(wr.Uid + ": ")
		buffer.WriteString(strconv.FormatInt(diff, 10))

		TestResults <- buffer.String()
	}
}


func TestPayload(wr *WorkRequest) {
	TestResults <- wr.Payload
}


func PrintlnEndpoint(wr *WorkRequest) {
	fmt.Println(wr.Payload)
}
