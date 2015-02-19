package main

import (
	"fmt"
	"time"
	"bytes"
	"strconv"
	"log"
	"net/url"
	"net/http"
	"crypto/tls"
_	"encoding/json"
	"github.com/anachronistic/apns"
//	"github.com/alexjlockwood/gcm" // No idea if this works
)

var (
	TestResults chan string
)

type endpoint func(wr *WorkRequest)


// *****************
// Testing endpoints
// *****************
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


// **********************
// Production endpoints
// **********************
func ApnsEndpoint(wr *WorkRequest) {
	payload := apns.NewPayload()
	payload.Alert = wr.Payload
	payload.Badge = 1
	payload.Sound = "bingbong.aiff"

	pn := apns.NewPushNotification()
	pn.DeviceToken = "969ed0884876465828f7945ee3141370b43e76a74179d3efdfbb726e80feeb36"
	pn.AddPayload(payload)

	// Create the client based on whether we are testing or not
	var client *apns.Client
	if (wr.Testing) {
		fmt.Println(config["apple_test_push_cert"], config["apple_test_push_cert"])
		client = apns.NewClient("gateway.sandbox.push.apple.com:2195",
			config["apple_push_test_cert"],
			config["apple_push_test_key"])
	} else {
		client = apns.NewClient("gateway.push.apple.com:2195",
			config["apple_push_cert"],
			config["apple_push_key"])
	}
	resp := client.Send(pn)

	alert, _ := pn.PayloadString()
	fmt.Println("  Token:", wr.Token)
	fmt.Println("  Alert:", alert)
	fmt.Println("Success:", resp.Success)
	fmt.Println("  Error:", resp.Error)
}

//GCM is totally borked right now
// func GcmEndpoint(wr *WorkRequest) {
// 	data := map[string]interface{}{"score": "5x1", "time": "15:10", "body": body}
// 	regIDs := []string{token}
// 	msg := gcm.NewMessage(data, regIDs...)

// 	sender := &gcm.Sender{ApiKey: config["gcm"]}

// 	response, err := sender.Send(msg, 2)
// 	if err != nil {
// 		fmt.Println("Failed to send message:", err)
// 	}
// 	fmt.Println(response)
// }



func WebsiteEndpoint(wr *WorkRequest) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	_, err := client.PostForm("https://192.241.199.78/event/",
		url.Values{"id": {wr.Token}, "payload": {wr.Payload}})
	if err != nil {
		log.Println(err)
	}
}
