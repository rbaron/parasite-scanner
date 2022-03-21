package main

import (
	"flag"
	"log"

	cfg "github.com/SysdigDan/parasite-scanner/pkg/config"
	"github.com/SysdigDan/parasite-scanner/pkg/data"
	"github.com/SysdigDan/parasite-scanner/pkg/exporter"
	"github.com/SysdigDan/parasite-scanner/pkg/mqtt"
)

var configFile = flag.String("config", "config.yaml", "YAML config filename")

func main() {
	flag.Parse()

	config, err := cfg.ParseConfig(*configFile)
	if err != nil {
		panic("unable to parse config file: " + err.Error())
	}

	dataSubscribers := []data.DataSubscriber{}
	if config.MQTT.Host != "" && config.MQTT.Enable {
		dataSubscribers = append(dataSubscribers, mqtt.MakeMQTTClient(&config.DEVICE, &config.MQTT))
	}
	if config.PROM.Host != "" && config.PROM.Enable {
		dataSubscribers = append(dataSubscribers, exporter.MakePROMExporter(&config.DEVICE, &config.PROM))
	}

	scanner := MakeParasiteScanner(&config.BLE)
	go scanner.Run()

	for _, subs := range dataSubscribers {
		go subs.Run()
	}

	for data := range scanner.channel {
		log.Println("[main] Got data:", data)
		for _, subs := range dataSubscribers {
			subs.Ingest(data)
		}
	}
}
