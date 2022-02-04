package main

import (
	"flag"
)

var showUI = flag.Bool("ui", false, "renders a terminal-based ui for iteractive use")
var configFile = flag.String("config", "config.yaml", "YAML config filename")

func main() {
	flag.Parse()

	config, err := ParseConfig(*configFile)
	if err != nil {
		panic("unable to parse config file: " + err.Error())
	}

	err = InitLogger(*showUI)
	if err != nil {
		panic("unable to initialize logger: " + err.Error())
	}
	defer DeInitLogger()

	dataSubscribers := []DataSubscriber{}
	if *showUI {
		dataSubscribers = append(dataSubscribers, InitUI())
	}
	if config.MQTT.Host != "" && config.MQTT.Enable {
		dataSubscribers = append(dataSubscribers, MakeMQTTClient(&config.DEVICE, &config.MQTT))
	}
	if config.PROM.Host != "" && config.PROM.Enable {
		dataSubscribers = append(dataSubscribers, MakePROMExporter(&config.DEVICE, &config.PROM))
	}

	scanner := MakeParasiteScanner(&config.BLE)
	go scanner.Run()

	for _, subs := range dataSubscribers {
		go subs.Run()
	}

	for data := range scanner.channel {
		logger.Println("[main] Got data:", data)
		for _, subs := range dataSubscribers {
			subs.Ingest(data)
		}
	}
}
