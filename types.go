package mailmanv2

import (
	apns "github.com/joekarl/go-libapns"
)

// An endpoint function
type endpoint func(*WorkRequest, *Worker)

// An error handling type for apns messages
type errorhandler func(*apns.Payload)

// A custom method function
type pushmethod func(*WorkRequest, *WorkRequest) *WorkRequest
