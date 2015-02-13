package main

import (
	"io/ioutil"
	"log"
	"strings"
)

// Config variables
var (
	configPath string = "mailman.conf"
	config map[string]interface{}
)

func InitConfig() {
	config = make(map[string]interface{})
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Println("Failed to open configfile: ", configPath)
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if len(line) > 0 {
			kv := strings.Split(line, "=")
			config[strings.Trim(kv[0], " ")] = strings.Trim(kv[1], " ")
		}
	}
}
