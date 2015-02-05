package main

import (
	"testing"
	"math/rand"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

func TestBulk(t *testing.T) {
	InitCollector(3)
	iterations := 10

	for i := 0 ; i < iterations ; i++ {
		// Random delay from 1 to 10 seconds
		delay := rand.Intn(10) + 1
		IssueWorkRequest(NewWorkRequest(strconv.Itoa(i), "Expected " + strconv.Itoa(delay) + "s", delay))
	}

	wg.Add(iterations)
	wg.Wait()
}
