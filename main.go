package main

import (
	"flag"
	"log"
	"os"
)

var showUI = flag.Bool("ui", false, "renders a terminal-based ui for iteractive use")

var tui *TUI
var logger *log.Logger

func main() {
	config, err := ParseConfig("example_config.yaml")
	if err != nil {
		panic("unable to parse config file " + err.Error())
	}

	flag.Parse()
	if *showUI {
		file, err := os.OpenFile("parasite-scanner.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("unable to open log file")
		}
		defer file.Close()
		logger = log.New(file, "", 0)
		tui = InitUI()
		go tui.Run()
	} else {
		logger = log.New(os.Stdout, "", 0)
	}

	scanner := MakeParasiteScanner(&config.BLE)
	mqttClient := MakeMQTTClient(&config.MQTT)

	go scanner.Run()
	go mqttClient.Run()

	for data := range scanner.channel {
		if tui != nil {
			tui.dataChan <- data
		}
		mqttClient.outgoing <- data
		logger.Println("[main] Got data:", data)
	}
}
