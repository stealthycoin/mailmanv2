package main

import (
	"testing"
	"math/rand"
	"strconv"
	"fmt"
	"time"
)

var (
	TestResults chan string
)

func TestBulk(t *testing.T) {
	// Seed rng and create a channel to recieve test results on
	rand.Seed(time.Now().UTC().UnixNano())
	TestResults = make(chan string)

	// Create a collector with three workers
	InitCollector(3)
	iterations := 1000 // Number of work requests to release

	for i := 0 ; i < iterations ; i++ {
		// Random delay from 1 to 10 seconds
		delay := rand.Intn(10) + 1
		IssueWorkRequest(NewWorkRequest(strconv.Itoa(i), "expected " + strconv.Itoa(delay) + "s", delay))
	}


	// Wait for the results and print them out for now
	for iterations > 0 {
		select {
		case result := <- TestResults:
			fmt.Println(result)
			iterations--
		}
	}
}
