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
	check_insert chan record_query  // Check if a record exists
)

type MailHeap []*mail_record

type record_query struct {
	record *mail_record
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
// Checks to see if a record exists and then inserts a record
//
func CheckAndInsertRecord(mr *mail_record) bool {
	respond := make(chan bool)
	check_insert <- record_query{mr, respond}
	result := <- respond
	return result
}


//
// Go routine for recording
//
func InitRecords(duration_fmt string) {
	log.SetFlags(log.Lshortfile)

	// Heap for timing
	h := &MailHeap{}
	heap.Init(h)
	counting := false

	// Hashmap for quick-check
	active_records := make(map[string]int)

	// Init channels
	file_record = make(chan *mail_record, 256)
	check_insert = make(chan record_query)
	clean := make(chan bool)

	// DUration
	dur, err := time.ParseDuration(duration_fmt)
	if err != nil {
		log.Fatal("Invalid record wait duration")
	}

	// ONLY HERE CALL GOOD, yes?
	FileRecord := func(record *mail_record) {
		// Put record into both datastructures
		log.Println("Filing record", record)
		heap.Push(h, record)
		if _, ok := active_records[record.Uid]; ok {
			active_records[record.Uid]++
		} else {
			active_records[record.Uid] = 1
		}

		if !counting {
			// If we aren't counting down start doing so now
			counting = true

			go func() {
				// Wait timelimit and then clean
				<- time.After(dur)
				clean <- true
			}()
		}
	}

	go func() {
		for {
			select {
			case record := <- file_record:
				FileRecord(record)

			case rq := <- check_insert:
				// Check if a record with the request id exists
				// return a bool through the rq.respond channel
				// and then add the record
				count, ok := active_records[rq.record.Uid]
				FileRecord(rq.record)

				rq.result <- ok && count > 0

			case <- clean:
				// Clean record storage
				log.Println("Cleaning")
				// Timer went off time to clean records
				counting = false
				done := false

				// Keep deleting entries until we find one that is too young
				for !done && h.Len() > 0 {
					// Get oldest record
					mr := heap.Pop(h).(*mail_record)
					now := time.Now().Unix()
					log.Println("\ttop record", mr, "age", time.Duration(now - mr.Last_alert) * time.Second)


					// check if it is from more than duration ago
					if time.Duration(now - mr.Last_alert) * time.Second >= dur {
						log.Println("\tpopping record")
						active_records[mr.Uid]--
						if active_records[mr.Uid] <= 0 {
							delete(active_records, mr.Uid)
						}
					} else {
						// This element is too young to delete but it is the oldest one
						// reset timer to when this needs to be deleted
						// and push it back into the heap
						heap.Push(h, mr)
						done = true
						counting = true
						go func() {
							<- time.After(dur - (time.Duration(now - mr.Last_alert) * time.Second))
							clean <- true
						}()
					}
				}
			}
		}
	}()
}
