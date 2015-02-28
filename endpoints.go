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
		diff -= time.Now().Unix()

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
// This doesnt do anything, send a messag with a duplicate hash and this set as the endpoint
// to effectivly cancel the previous message wiht that hash
func CancelEndpoint(wr *WorkRequest) {}

func PhoneEndpoint(wr *WorkRequest) {
	ApnsEndpoint(wr) // Suave
}

func ApnsEndpoint(wr *WorkRequest) {
	payload := apns.NewPayload()
	payload.Alert = wr.Payload
	payload.Badge = 1
	payload.Sound = "bingbong.aiff"

	pn := apns.NewPushNotification()
	// Getting the device token is a matter of fetching a lot of stuff from the database make this one query later
	var id int64
	err := db.QueryRow(`select user_id from main_userprofile where id = $1`, wr.Token).Scan(&id)
	if err != nil {
		log.Println(err)
		return
	}
	err = db.QueryRow(`select registration_id from push_notifications_apnsdevice where user_id = $1`, id).Scan(&pn.DeviceToken)
	if err != nil {
		log.Println(err)
		return
	}
	pn.AddPayload(payload)

	// Create the client based on whether we are testing or not
	var client *apns.Client
	wr.Testing = true
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
	// Ignoring errors like a good boi
	client.Send(pn)

	pn.PayloadString()
	// fmt.Println("  Token:", wr.Token)
	// fmt.Println("  Alert:", alert)
	// fmt.Println("Success:", resp.Success)
	// fmt.Println("  Error:", resp.Error)
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
