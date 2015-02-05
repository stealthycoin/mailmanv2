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
	rand.Seed(time.Now().UTC().UnixNano())
	TestResults = make(chan string)

	InitCollector(3)
	iterations := 10

	for i := 0 ; i < iterations ; i++ {
		// Random delay from 1 to 10 seconds
		delay := rand.Intn(10) + 1
		IssueWorkRequest(NewWorkRequest(strconv.Itoa(i), "Expected " + strconv.Itoa(delay) + "s", delay))
	}


	for iterations > 0 {
		select {
		case result := <- TestResults:
			fmt.Println(result)
			iterations--
		}
	}
}
