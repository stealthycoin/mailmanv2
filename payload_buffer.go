package mailmanv2

import (
	apns "github.com/joekarl/go-libapns"
)

//
// Buffer to hold sent payloads for error handling
//
type PayloadBuffer struct {
	buffer        []*apns.Payload
	buffer_offset uint32
	error         bool
}
