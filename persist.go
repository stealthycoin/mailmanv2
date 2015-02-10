package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"encoding/gob"
)

var (
	backupPath = "requests.gob"
)

// Launch persistence
func InitPersist() {
	// Gob type registration
	gob.Register(requests)

	// Handle the HUP signal which will be defined to backup memory
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		for {
			<- hup
			BackupRequests()
		}
	}()

	// Handle the INT/TERM signal which should also back up memory
	sigs := make (chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<- sigs
		BackupRequests()
		os.Exit(1)
	}()

	LoadRequests()
}

func LoadRequests() {
	// Read gob into requests
	file, err := os.Open(backupPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(&requests)
	if err != nil {
		fmt.Println(err, "Cannot decode file")
	}

	// Reload all work requests
	for _, req := range requests {
		go req.StartTimer()
	}
}

func BackupRequests() {
	// Check if backup file already exists
	file, err := os.Create(backupPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	enc.Encode(requests)
}
