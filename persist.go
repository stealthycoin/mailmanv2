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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for {
			<- sigs
			BackupRequests()
		}
	}()
}

func LoadRequests() {
	file, err := os.Open(backupPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	dec.Decode(requests)
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
