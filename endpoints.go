package main

import (
	"time"
	"bytes"
	"strconv"
)

type endpoint func(wr *WorkRequest)

func TestTimePayload(wr *WorkRequest) {
	var buffer bytes.Buffer

	// Write testing message
	diff, err := strconv.ParseInt(wr.payload, 10, 64)
	if err != nil {
		TestResults <- "Oh fuck"
	} else {
		diff -= time.Now().UnixNano()
		diff /= 1000000000

		buffer.WriteString(wr.uid + ": ")
		buffer.WriteString(strconv.FormatInt(diff, 10))

		TestResults <- buffer.String()
	}
}


func TestPayload(wr *WorkRequest) {
	TestResults <- wr.payload
}
