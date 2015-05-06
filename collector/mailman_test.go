package collector

import (
	"testing"
	"math/rand"
	"strconv"
	"fmt"
	"time"
	"regexp"
)


func init() {
	InitPersist()
	StartRecords("2s")
}


func TestBulk(t *testing.T) {
	// Seed rng and create a channel to recieve test results on
	rand.Seed(time.Now().Unix())
	TestResults = make(chan string)

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
	CollectRequest <- NewWorkRequest("ID", "cancel", "add", "0", "n/a", 0) // Cancel the message coming in 1 second
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
