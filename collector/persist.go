package collector

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"encoding/gob"
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
	file, err := os.Open(config["backup_path"])
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(&requests)
	if err != nil {
		log.Println("Cannot decode backup file")
	}

	// Reload all work requests
	for _, req := range requests {
		req.Cancel = make(chan bool)
		go req.StartTimer()
	}
}

func BackupRequests() {
	// Check if backup file already exists
	file, err := os.Create(config["backup_path"])
	if err != nil {
		log.Println(err, config["backup_path"])
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	enc.Encode(requests)
}
