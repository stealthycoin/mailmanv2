package collector

import (
	"time"
	"container/heap"
)

// This package prevents people from getting too many beeps on their phone
// It maintains records of recently sent messages, if a record exists
// future messages should not beep, probably.
var (
	file_record chan *mail_record
	check_record chan record_query  // Check if a record exists
)

type MailHeap []mail_record

type record_query struct {
	id int64
	result chan bool
}

type mail_record struct {
	Uid int64          // User ID
	Last_alert int64   // Last time their were rang
}

//
// Heap interface fns
//
func (h MailHeap) Len() int {
	return len(h)
}

func (h MailHeap) Less(i, j int) bool {
	return h[i].Last_alert < h[j].Last_alert // Oldest record at the top
}

func (h MailHeap) Swap(i, j int ) {
	h[i], h[j] = h[j], h[i]
}

func (h *MailHeap) Push(x interface{}) {
	*h = append(*h, x.(mail_record))
}

func (h *MailHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0:n-1]
	return x
}


//
// Checks to see if a record exists
//
func CheckRecord(id int64) bool {
	respond := make(chan bool)
	check_record <- record_query{id, respond}
	if result := <- respond; result {
		return true
	}
	return false
}


//
// Go routine for recording
//
func StartRecords() {
	// Heap for timing
	h := &MailHeap{}
	heap.Init(h)
	var timer time.Timer
	counting := false

	// Hashmap for quick-check
	active_records := make(map[int64]int)

	// Init channels
	file_record = make(chan *mail_record, 128)
	check_record = make(chan record_query)

	go func() {
		for {
			select {
			case record := <- file_record:
				// Put record into both datastructures
				h.Push(record)
				if _, ok := active_records[record.Uid]; ok {
					active_records[record.Uid]++
				} else {
					active_records[record.Uid] = 1
				}

				if !counting {
					// If we aren't counting down start doing so now
					timer.Reset(time.Minute * 5)
					counting = true
				}


			case <- timer.C:
				// Timer went off time to delete elements
				now := time.Now().Unix()
				done := false

				// Keep deleting entries until we find one that is too young
				for !done && h.Len() > 0 {
					mr := (*h)[0]

					// Check if it is from more than 5 minutes ago
					if now - mr.Last_alert > 5 * 60 {
						h.Pop() // if its older than 5 minutes pop it
						active_records[mr.Uid]--
						if active_records[mr.Uid] <= 0 {
							delete(active_records, mr.Uid)
						}
					}

					// This element is too young to delete but it is the oldest one
					// Reset timer to when this needs to be deleted
					delay := (5 * 60) - (now - mr.Last_alert)
					timer.Reset(time.Duration(delay) * time.Second)
				}
			}
		}
	}()
}
