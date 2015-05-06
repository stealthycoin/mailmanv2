package collector

import (
	"log"
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

type MailHeap []*mail_record

type record_query struct {
	id string
	result chan bool
}

type mail_record struct {
	Uid string         // User ID
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
	*h = append(*h, x.(*mail_record))
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
func CheckRecord(id string) bool {
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
func StartRecords(duration_fmt string) {
	log.SetFlags(log.Lshortfile)

	// Heap for timing
	h := &MailHeap{}
	heap.Init(h)
	counting := false

	// Hashmap for quick-check
	active_records := make(map[string]int)

	// Init channels
	file_record = make(chan *mail_record, 256)
	check_record = make(chan record_query)
	clean := make(chan bool)

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
					counting = true

					go func() {
						// Wait 5 minutes and then clean
						if dur, err := time.ParseDuration(duration_fmt); err == nil {
							<- time.After(dur)
							clean <- true
						} else {
							log.Fatal(err)
						}
					}()
				}
			case rq := <- check_record:
				// Check if a record exists
				if count, ok := active_records[rq.id]; ok && count > 0{
					rq.result <- true
				}
				rq.result <- false

			case <- clean:
				// Clean record storage
				log.Println("Cleaning")
				// Timer went off time to clean records
				counting = false
				now := time.Now().Unix()
				done := false

				// Keep deleting entries until we find one that is too young
				for !done && h.Len() > 0 {
					// Get oldest record
					mr := (*h)[0]

					// Check if it is from more than 5 minutes ago
					if now - mr.Last_alert > 5 * 60 {
						h.Pop() // if its older than 5 minutes pop it
						active_records[mr.Uid]--
						if active_records[mr.Uid] <= 0 {
							delete(active_records, mr.Uid)
						}
					} else {
						// This element is too young to delete but it is the oldest one
						// Reset timer to when this needs to be deleted
						done = true
						counting = true
						go func() {
							delay, err := time.ParseDuration(duration_fmt)
							if err != nil {
								log.Fatal(err)
							}
							delay_secs := delay.Seconds()  - float64(now - mr.Last_alert)

							<- time.After(time.Duration(delay_secs) * time.Second)
							clean <- true
						}()
					}
				}
			}
		}
	}()
}
