package main

import (
	"fmt"
	"time"
	"bytes"
	"strconv"
	"github.com/anachronistic/apns"
	"github.com/alexjlockwood/gcm" // No idea if this works
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


// Production endpoints
func ApnsEndpoint(token, body string) {
	payload := apns.NewPayload()
	payload.Alert = "Hello, World!"
	payload.Badge = 42
	payload.Sound = "bingbong.aiff"

	pn := apns.NewPushNotification()
	pn.DeviceToken = token
	pn.AddPayload(payload)

	client := apns.NewClient("gateway.push.apple.com:2195",
		"/home/theupdatesmen/go/src/bitbucket.org/push_server/certs/HearthPushCert.pem",
		"/home/theupdatesmen/go/src/bitbucket.org/push_server/certs/HearthPushKey.pem")
	resp := client.Send(pn)
	alert, _ := pn.PayloadString()
	fmt.Println("  Token:", token)
	fmt.Println("  Alert:", alert)
	fmt.Println("Success:", resp.Success)
	fmt.Println("  Error:", resp.Error)
}


func GcmEndpoint(token, body string) {
	data := map[string]interface{}{"score": "5x1", "time": "15:10", "body": body}
	regIDs := []string{token}
	msg := gcm.NewMessage(data, regIDs...)

	sender := &gcm.Sender{ApiKey: config["gcm"].(string)}

	response, err := sender.Send(msg, 2)
	if err != nil {
		fmt.Println("Failed to send message:", err)
	}
	fmt.Println(response)
}
