package main

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
}


func TestBulk(t *testing.T) {
	// Seed rng and create a channel to recieve test results on
	rand.Seed(time.Now().UnixNano())
	TestResults = make(chan string)

	// Create a collector with three workers
	InitCollector(3)
	iterations := 10000 // Number of work requests to release

	for i := 0 ; i < iterations ; i++ {
		// Random delay from 1 to 10 seconds
		delay := rand.Intn(10) + 1

		target := time.Now().UnixNano() + int64(delay) * 1000000000
		IssueWorkRequest(NewWorkRequest(strconv.Itoa(i), "testtime", strconv.FormatInt(target, 10), target))
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
	IssueWorkRequest(NewWorkRequest("ID", "testpayload", "message1", time.Now().UnixNano() + 1000000000))
	IssueWorkRequest(NewWorkRequest("ID", "testpayload", "message2", time.Now().UnixNano() + 2000000000))

	result := <- TestResults
	if result == "message1" {
		t.Errorf("Wrong message\n")
	}
	StopCollector()
}

func TestBackup(t *testing.T) {
	InitCollector(1)
	IssueWorkRequest(NewWorkRequest("ID", "testpayload", "TURTLE POWER", time.Now().UnixNano() + 3000000000))
	BackupRequests()
	LoadRequests()
	if (requests["ID"].Payload != "TURTLE POWER") {
		t.Errorf("Loaded value does not match saved value")
	}
	StopCollector()
}
