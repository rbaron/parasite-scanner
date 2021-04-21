package main

import (
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

var db = map[string][]ParasiteData{}

type ParasiteData struct {
	Counter        uint8
	BatteryVoltage float32
	TempCelcius    float32
	Humidity       float32
	SoilMoisture   float32
	Time           time.Time
}

func (pd ParasiteData) String() string {
	return fmt.Sprintf("Counter %d; BatteryVoltage: %f", pd.Counter, pd.BatteryVoltage)
}

func ParseParasiteeData(serviceData bluetooth.AdvServiceData) ParasiteData {
	// if serviceData.UUID.Bytes()[0] != 0x18 || serviceData.UUID.Bytes()[1] != 0x1a {
	// 	fmt.Println("Unable to parse!")
	// 	return ParasiteData{}
	// }

	data := serviceData.Data

	counter := data[1] & 0x0f
	batteryVoltage := binary.BigEndian.Uint16(data[2:4])
	tempCelcius := binary.BigEndian.Uint16(data[4:6])
	humidity := binary.BigEndian.Uint16(data[6:8])
	soilMoisture := binary.BigEndian.Uint16(data[8:10])

	return ParasiteData{
		Counter:        counter,
		BatteryVoltage: float32(batteryVoltage) / 1000,
		TempCelcius:    float32(tempCelcius) / 1000,
		Humidity:       float32(humidity) / (1 << 16),
		SoilMoisture:   100 * float32(soilMoisture) / (1 << 16),
		Time:           time.Now(),
	}
}

func DumpDB() {
	keys := make([]string, 0)
	for k, _ := range db {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		var last = db[k][len(db[k])-1]
		fmt.Printf(
			"%s | soil: %5.1f%% | batt: %3.1fV | temp: %4.1fC | humi: %5.1f%% | %6.1fs ago\n",
			k,
			last.SoilMoisture,
			last.BatteryVoltage,
			last.TempCelcius,
			last.Humidity,
			time.Since(last.Time).Seconds())
	}
	fmt.Println("-----------------------------")
}

func main() {
	if err := adapter.Enable(); err != nil {
		panic("Unable to initialize the BLE stack: " + err.Error())
	}

	println("Waiting for parasites' BLE advertisements...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.LocalName() == "prst" {
			key := device.Address.String()
			data := ParseParasiteeData(device.AdvertisementPayload.GetServiceDatas()[0])

			values, exists := db[key]
			if !exists {
				db[key] = make([]ParasiteData, 10)
				DumpDB()
			} else if values[len(values)-1].Counter != data.Counter {
				db[key] = append(db[key], data)
				DumpDB()
			}
		}
	})

	if err != nil {
		panic("Unable to start scanning: " + err.Error())
	}
}
