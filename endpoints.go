package main

import (
	"time"
	"bytes"
	"strconv"
)

func TestPayload(wr * WorkRequest) {
	var buffer bytes.Buffer
	buffer.WriteString(wr.uid + ": took ")
	buffer.WriteString(strconv.FormatInt((time.Now().UnixNano() - wr.start_time) / int64(time.Second), 10))
	buffer.WriteString("s " + wr.payload)
	TestResults <- buffer.String()
}
