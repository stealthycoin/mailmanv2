package collector

import (
	"io/ioutil"
	"log"
	"os"
	"time"
	"os/signal"
	"syscall"
	"strings"
	"strconv"
	"github.com/anachronistic/apns"
)

// Config variables
var (
	configPath string = "mailman.conf"
	Config map[string]string

	// APNS client
	apns_test, apns_real *apns.Client
)

func InitConfig() {
	Config = make(map[string]string)

	// Listen for USR1 signal to reload the conf file
	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGUSR1)
	go func() {
		for {
			<- reload
			StopCollector()
			LoadConfig()
			wc, _ := strconv.Atoi(Config["workers"])
			InitCollector(wc)
		}
	}()

	LoadConfig()

	// Backup every td seconds
	go func() {
		for {
			td := 3600 // Default timedelay to an hour
			if delay, ok := Config["backup_delay"] ; ok {
				td, _ = strconv.Atoi(delay)
			}
			time.Sleep(time.Duration(td) * time.Second)
			backup <- true
		}
	}()
}

func LoadConfig() {
	log.Println("Loading Config.")
	// Defaults set here and will be overrided by the Config file
	Config["backup_path"] = "backup.gob"

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Println("Failed to open Config file: ", configPath)
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.Trim(line, " ")
		if i := strings.Index(line, "#") ; i >= 0 {
			line = line[0:i]
		}
		line = strings.Trim(line, " ")
		if len(line) > 0 {
			kv := strings.Split(line, "=")
			Config[strings.Trim(kv[0], " ")] = strings.Trim(kv[1], " ")
		}
	}

	// Once config is loaded load the apns client
	apns_test = apns.NewClient("gateway.sandbox.push.apple.com:2195",
		Config["apple_push_test_cert"],
		Config["apple_push_test_key"])
	apns_real = apns.NewClient("gateway.push.apple.com:2195",
		Config["apple_push_cert"],
		Config["apple_push_key"])
}
