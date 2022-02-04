package main

import (
	"fmt"
	"time"
)

// ParasiteData is the main currency in parasite-scanner.
// The BLE scanner listens for b-parasite broadcasts and instantiate a ParasiteData
// object whenever a valid message is received, after deduplication.
// Consumers of ParasiteData should implement the DataSubscriber interface below, and
// will be fed new data upon arrival.
type ParasiteData struct {
	Key               string
	Counter           uint8
	BatteryPercentage float64
	BatteryVoltage    float64
	TempCelcius       float64
	Humidity          float64
	Time              time.Time
	RSSI              int
}

func (pd ParasiteData) String() string {
	return fmt.Sprintf(
		"%s | battp: %2.1f%% | battv: %3.1fV | temp: %4.1fC | humi: %5.1f%% | %6.1fs ago | counter: %d",
		pd.Key,
		pd.BatteryPercentage,
		pd.BatteryVoltage,
		pd.TempCelcius,
		pd.Humidity,
		time.Since(pd.Time).Seconds(),
		pd.Counter)
}

type DataSubscriber interface {
	// A blocking function that will be called on its own go routine.
	Run()
	// A function that will be called whenever new data is available.
	Ingest(data *ParasiteData)
}
