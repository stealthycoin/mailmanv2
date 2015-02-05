package main

import (
	"testing"
	"math/rand"
	"strconv"
	"fmt"
	"time"
	"regexp"
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
	errors := 0
	for iterations > 0 {
		select {
		case result := <- TestResults:
			re, _ := regexp.Compile("([0-9]+): took ([0-9]+)s expected ([0-9]+)")
			match := re.FindStringSubmatch(result)
			if match[2] != match[3] {
				errors++
			}
			iterations--
		}
	}
	if errors > 0 {
		t.Errorf("%d message(s) not delivered on time.\n", errors)
	} else {
		fmt.Printf("All messages were delivered on time.\n")
	}
}
