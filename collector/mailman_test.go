package collector

import (
	"testing"
	"math/rand"
	"strconv"
	"fmt"
	"log"
	"time"
	"regexp"
	_ "github.com/lib/pq"
	"database/sql"
)


func init() {
	TestResults = make(chan string)
	InitConfig()
	InitPersist()
	StartRecords("2s")
}


func TestBulk(t *testing.T) {
	// Seed rng and create a channel to recieve test results on
	rand.Seed(time.Now().Unix())

	// Create a collector with three workers
	InitCollector(3)
	iterations := 10000 // Number of work requests to release

	for i := 0 ; i < iterations ; i++ {
		// Random delay from 1 to 10 seconds
		delay := rand.Intn(10) + 1

		target := time.Now().Unix() + int64(delay) * 1
		CollectRequest <- NewWorkRequest(strconv.Itoa(i), "testtime", "add", "", strconv.FormatInt(target, 10), target)
	}

	// Wait for the results and print them out for now
	errors := 0
	re, _ := regexp.Compile("[0-9]+: ([0-9]+)")
	for iterations > 0 {
		result := <- TestResults
		match := re.FindStringSubmatch(result)
		if len(match) != 2 || match[1] != "0" {
			fmt.Println(result)
			errors++
		}
		iterations--
	}
	if errors > 0 {
		t.Errorf("%d message(s) not delivered on time.\n", errors)
	} else {
		fmt.Printf("All messages were delivered on time.\n")
	}

	StopCollector()
}


func TestReplace(t *testing.T) {
	InitCollector(2)
	CollectRequest <- NewWorkRequest("ID", "testpayload", "add", "0", "message1", time.Now().Unix() + 1)
	CollectRequest <- NewWorkRequest("ID", "testpayload", "add", "0", "message2", time.Now().Unix() + 2)

	result := <- TestResults
	if result == "message1" {
		t.Errorf("Wrong message\n")
	}
	StopCollector()
}

func TestBackup(t *testing.T) {
	InitCollector(1)
	CollectRequest <- NewWorkRequest("ID", "testpayload", "add", "", "TURTLE POWER", time.Now().Unix() + 3)
	BackupRequests()
	LoadRequests()
	if (requests["ID"].Payload != "TURTLE POWER") {
		t.Errorf("Loaded value does not match saved value")
	}
	StopCollector()
}

// Deliver a webiste notification in 5 seconds
func TestWebsiteDelivery(t *testing.T) {
	InitCollector(1)

	CollectRequest <- NewWorkRequest("ID", "website", "add", "3", `{"title":"You have earned the Selfie badge", "img": "https://www.hearthapp.net/static/images/badge/selfie.png", "imgwidth": 60, "content": "Upload a profile picture for the first time."}`, time.Now().Unix() + 5)
	time.Sleep(time.Duration(6) * time.Second)

	StopCollector()
}


func TestCancel(t *testing.T) {
	InitCollector(1)
	fmt.Println("Testing Cancel")
	CollectRequest <- NewWorkRequest("ID", "testpayload", "add", "0", "message", time.Now().Unix() + 1) // Deliver in 1 seconds
	CollectRequest <- NewWorkRequest("ID", "", "cancel", "0", "n/a", 0) // Cancel the message coming in 1 second
	CollectRequest <- NewWorkRequest("ID2", "testpayload", "add", "0", "yay", time.Now().Unix() + 2) // Deliver in 2 seconds

	// Should get yay since message was canceled
	result := <- TestResults
	if result != "yay" {
		t.Errorf("Wrong message\n")
	} else {
		fmt.Println("Cancel successful")
	}
	StopCollector()
}

func TestPushNote(t *testing.T) {
	InitCollector(5)
	// Who to send messages to, user pk in local db
	user_id := "3"
	fmt.Println("Testing Push Notifications on userid:", user_id)

	// Get a database connection
	tdb, err := sql.Open("postgres", "dbname=hearth user=hearth host=localhost password=A938CEA3C22F8FD93F4157D4A1AB3AF753452D743FEC6A8B27401972B3F9511F sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	SetDb(tdb)

	// Send 3 batches of 3 notifications 3 seconds apart
	// Checking to see if the silencing mechanism works as
	// well as the push notifications themselves
	// records is set to release a record after 2 seconds
	// so 3 should be enough time to unsilence the beeping

	// Lame way to make a bunch of things
	// Its this way so if I want to change it to not be a pattern
	// later I don't have to undo a weird loop and write them all out
	payloada := `{"message":"TEST: Batch 1 MSG 1"}`
	payloadb := `{"message":"TEST: Batch 1 MSG 2"}`
	payloadc := `{"message":"TEST: Batch 1 MSG 3"}`

	payloadd := `{"message":"TEST: Batch 2 MSG 1"}`
	payloade := `{"message":"TEST: Batch 2 MSG 2"}`
	payloadf := `{"message":"TEST: Batch 2 MSG 3"}`

	payloadg := `{"message":"TEST: Batch 3 MSG 1"}`
	payloadh := `{"message":"TEST: Batch 3 MSG 2"}`
	payloadi := `{"message":"TEST: Batch 3 MSG 3"}`

	now := time.Now().Unix()

	CollectRequest <- NewWorkRequest("ID1", "phone", "add", user_id, payloada, 0)
	CollectRequest <- NewWorkRequest("ID2", "phone", "add", user_id, payloadb, 0)
	CollectRequest <- NewWorkRequest("ID3", "phone", "add", user_id, payloadc, 0)

	CollectRequest <- NewWorkRequest("ID4", "phone", "add", user_id, payloadd, now + 3)
	CollectRequest <- NewWorkRequest("ID5", "phone", "add", user_id, payloade, now + 3)
	CollectRequest <- NewWorkRequest("ID6", "phone", "add", user_id, payloadf, now + 3)

	CollectRequest <- NewWorkRequest("ID7", "phone", "add", user_id, payloadg, now + 6)
	CollectRequest <- NewWorkRequest("ID8", "phone", "add", user_id, payloadh, now + 6)
	CollectRequest <- NewWorkRequest("ID9", "phone", "add", user_id, payloadi, now + 6)
	<- time.After(10 * time.Second)
	StopCollector()
}
