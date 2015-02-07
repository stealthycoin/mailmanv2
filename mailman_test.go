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

func TestBulk(b *testing.T) {
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
		IssueWorkRequest(NewWorkRequest(strconv.Itoa(i), strconv.FormatInt(target, 10), target))
	}

	// Wait for the results and print them out for now
	errors := 0
	re, _ := regexp.Compile("[0-9]+: ([0-9]+)")
	for iterations > 0 {
		result := <- TestResults
		match := re.FindStringSubmatch(result)
		if match[1] != "0" {
			fmt.Println(result)
			errors++
		}
		iterations--
	}
	if errors > 0 {
		b.Errorf("%d message(s) not delivered on time.\n", errors)
	} else {
		fmt.Printf("All messages were delivered on time.\n")
	}
}
