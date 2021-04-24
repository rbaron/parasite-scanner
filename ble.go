package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

type ParasiteData struct {
	Key            string
	Counter        uint8
	BatteryVoltage float32
	TempCelcius    float32
	Humidity       float32
	SoilMoisture   float32
	Time           time.Time
	RSSI           int
}

func (pd ParasiteData) String() string {
	return fmt.Sprintf(
		"%s | soil: %5.1f%% | batt: %3.1fV | temp: %4.1fC | humi: %5.1f%% | %6.1fs ago\n",
		pd.Key,
		pd.SoilMoisture,
		pd.BatteryVoltage,
		pd.TempCelcius,
		pd.Humidity,
		time.Since(pd.Time).Seconds())
}

func parseParasiteData(scanResult bluetooth.ScanResult) (ParasiteData, error) {
	if len(scanResult.AdvertisementPayload.GetServiceDatas()) != 1 {
		return ParasiteData{}, fmt.Errorf("unexpected length of service datas")
	}

	serviceData := scanResult.AdvertisementPayload.GetServiceDatas()[0]

	uuid := serviceData.UUID.Bytes()
	if uuid[len(uuid)-1] != 0x18 || uuid[len(uuid)-2] != 0x1a {
		return ParasiteData{}, fmt.Errorf("invalid service data uuid: %s", uuid)
	}

	counter := serviceData.Data[1] & 0x0f
	batteryVoltage := binary.BigEndian.Uint16(serviceData.Data[2:4])
	tempCelcius := binary.BigEndian.Uint16(serviceData.Data[4:6])
	humidity := binary.BigEndian.Uint16(serviceData.Data[6:8])
	soilMoisture := binary.BigEndian.Uint16(serviceData.Data[8:10])

	return ParasiteData{
		Key:            scanResult.Address.String(),
		Counter:        counter,
		BatteryVoltage: float32(batteryVoltage) / 1000,
		TempCelcius:    float32(tempCelcius) / 1000,
		Humidity:       100 * float32(humidity) / (1 << 16),
		SoilMoisture:   100 * float32(soilMoisture) / (1 << 16),
		Time:           time.Now(),
		RSSI:           int(scanResult.RSSI),
	}, nil
}

func StartScanning(ch chan ParasiteData) {
	var adapter = bluetooth.DefaultAdapter

	if err := adapter.Enable(); err != nil {
		panic("Unable to initialize the BLE stack: " + err.Error())
	}

	err := adapter.Scan(func(adapter *bluetooth.Adapter, scanResult bluetooth.ScanResult) {
		if scanResult.LocalName() == "prst" {
			data, err := parseParasiteData(scanResult)
			if err != nil {
				fmt.Println("error parsing parasite data:", err.Error())
			} else {
				ch <- data
			}
		}
	})

	if err != nil {
		panic("unable to start scanning: " + err.Error())
	}

}
