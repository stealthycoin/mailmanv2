package collector

import (
	"fmt"
	"time"
	"bytes"
	"strconv"
	"log"
	"net/url"
	"net/http"
	"crypto/tls"
	"encoding/json"
	"database/sql"
	"github.com/anachronistic/apns"
	"github.com/alexjlockwood/gcm"
)

var (
	db *sql.DB
	TestResults chan string
)

type endpoint func(wr *WorkRequest)


func SetDb(newdb *sql.DB) {
	db = newdb
}

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
	// First we need to fetch user profile
	var id int64
	err := db.QueryRow(`select user_id from main_userprofile where id = $1`, wr.Token).Scan(&id)
	if err != nil {
		log.Println(err)
		return
	}

	// Fetch all the apns devices
	apns_devices := make([]Phone,0,0)
	rows, err := db.Query(`select registration_id, name from push_notifications_apnsdevice
                           where user_id = $1 and active = TRUE`, id)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		d := Phone{"","apple",""}
		err := rows.Scan(&d.reg_id, &d.name)
		if err != nil {
			log.Println(err)
			return
		}
		apns_devices = append(apns_devices, d)
	}

	// Fetch all gcm devices
	gcm_devices := make([]Phone,0,0)
	rows, err = db.Query(`select registration_id, name from push_notifications_gcmdevice
                           where user_id = $1 and active = TRUE`, id)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		d := Phone{"","android",""}
		err := rows.Scan(&d.reg_id, &d.name)
		if err != nil {
			log.Println(err)
			return
		}
		gcm_devices = append(gcm_devices, d)
	}


	// Send apns messages
	for _, device := range apns_devices {
		ApnsEndpoint(device, wr) // Suave
	}

	// Send gcm messages
	for _, device := range gcm_devices {
		GcmEndpoint(device, wr)
	}
}

func ApnsEndpoint(device Phone, wr *WorkRequest) {

	// Unmarshal the payload
	var dict map[string]interface{}
	err := json.Unmarshal([]byte(wr.Payload), &dict)
	if err != nil {
		log.Println(err)
		return
	}


	payload := apns.NewPayload()
	payload.Alert = dict["message"].(string)
	payload.Badge = 1
	payload.Sound = "default"

	pn := apns.NewPushNotification()
	// Getting the device token is a matter of fetching a lot of stuff from the database make this one query later
	var id int64
	err = db.QueryRow(`select user_id from main_userprofile where id = $1`, wr.Token).Scan(&id)
	if err != nil {
		log.Println(err)
		return
	}
	var testing string
	err = db.QueryRow(`select registration_id, name from push_notifications_apnsdevice
                       where user_id = $1 and active = TRUE`, id).Scan(&pn.DeviceToken, &testing)
	if err != nil {
		log.Println(err)
		return
	}
	pn.AddPayload(payload)
	// Add custom keys to the pn
	for key, val := range dict {
		if key != "message" { // Don't copy the message twice since we are sending it in Alert
			pn.Set(key, val)
		}
	}

	// Create the client based on whether we are testing or not
	var client *apns.Client

	if testing == "testing" {
		fmt.Println(Config["apple_test_push_cert"], Config["apple_test_push_cert"])
		client = apns.NewClient("gateway.sandbox.push.apple.com:2195",
			Config["apple_push_test_cert"],
			Config["apple_push_test_key"])
	} else {
		client = apns.NewClient("gateway.push.apple.com:2195",
			Config["apple_push_cert"],
			Config["apple_push_key"])
	}
	// Ignoring errors like a good boi
	client.Send(pn)

	// pn.PayloadString()
	// fmt.Println("  Token:", wr.Token)
	// fmt.Println("  Alert:", alert)
	// fmt.Println("Success:", resp.Success)
	// fmt.Println("  Error:", resp.Error)
}

// GCM is working!
func GcmEndpoint(device Phone, wr *WorkRequest) {
	data := map[string]interface{}{"title":"Hearth","message": wr.Payload,"msgcnt":1, "notId": time.Now().Unix()}
	regIDs := []string{device.reg_id}
	msg := gcm.NewMessage(data, regIDs...)

	sender := &gcm.Sender{ApiKey: Config["gcm"]}

	response, err := sender.Send(msg, 2)
	if err != nil {
		fmt.Println("Failed to send message:", err)
	}
	fmt.Println(response)
}



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
