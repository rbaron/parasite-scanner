package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const kMacOSMACAddrPrefix = "f0:ca:f0:ca:"

type ParasiteScanner struct {
	// We use a wrap-around counter inside the advertisement payload for
	// deduplicating messages.
	lastCounter map[string]int
	channel     chan *ParasiteData
	cfg         *BLEConfig
}

func MakeParasiteScanner(cfg *BLEConfig) *ParasiteScanner {
	return &ParasiteScanner{
		lastCounter: map[string]int{},
		channel:     make(chan *ParasiteData),
		cfg:         cfg,
	}
}

// This is a workaround for getting the MAC address on macOS.
// Ideally, we'd like the "key" for every scan result to be the
// MAC address for the discovered device.
// On Linux, that works well and that's what we use.
// On macOS, scanned devices' MAC addresses are hidden for privacy
// reasons, and the API only returns an UUID for us instead.
// To get around this, I encoded the last two bytes of parasites'
// MAC addresses in their advertisement data, so we can sort of
// infer the whole MAC address if we assume the leading bytes.
func getKey(cfg *BLEConfig, scanResult *bluetooth.ScanResult) string {
	addr := scanResult.Address.String()
	if !cfg.MacOS.InferMACAddress {
		return addr
	}

	isUUID := len(addr) == 36

	// // Linux - we already have the MAC address, so just return that.
	if !isUUID {
		return addr
	} else {
		// macOS - we try to read the MAC address from the payload data.
		serviceData := scanResult.AdvertisementPayload.GetServiceDatas()[0].Data
		if len(serviceData) >= 12 {
			lsb1 := serviceData[10]
			lsb0 := serviceData[11]
			return fmt.Sprintf("%s:%02d:%02d", cfg.MacOS.MACAddressPrefix, lsb1, lsb0)
		}
		return addr
	}
}

func parseParasiteData(cfg *BLEConfig, scanResult bluetooth.ScanResult) (*ParasiteData, error) {
	if len(scanResult.AdvertisementPayload.GetServiceDatas()) != 1 {
		return nil, fmt.Errorf("unexpected length of service datas")
	}

	serviceData := scanResult.AdvertisementPayload.GetServiceDatas()[0]

	uuid := serviceData.UUID
	if !uuid.Is16Bit() || uuid.Get16Bit() != 0x181a {
		return nil, fmt.Errorf("invalid service data uuid: %s", uuid)
	}

	counter := serviceData.Data[1] & 0x0f
	batteryVoltage := binary.BigEndian.Uint16(serviceData.Data[2:4])
	tempCelcius := binary.BigEndian.Uint16(serviceData.Data[4:6])
	humidity := binary.BigEndian.Uint16(serviceData.Data[6:8])
	soilMoisture := binary.BigEndian.Uint16(serviceData.Data[8:10])

	return &ParasiteData{
		Key:            getKey(cfg, &scanResult),
		Counter:        counter,
		BatteryVoltage: float32(batteryVoltage) / 1000,
		TempCelcius:    float32(tempCelcius) / 1000,
		Humidity:       100 * float32(humidity) / (1 << 16),
		SoilMoisture:   100 * float32(soilMoisture) / (1 << 16),
		Time:           time.Now(),
		RSSI:           int(scanResult.RSSI),
	}, nil
}

func (scanner *ParasiteScanner) Run() {
	var adapter = bluetooth.DefaultAdapter

	if err := adapter.Enable(); err != nil {
		panic("unable to initialize the BLE stack: " + err.Error())
	}

	err := adapter.Scan(func(adapter *bluetooth.Adapter, scanResult bluetooth.ScanResult) {
		if scanResult.LocalName() == "prst" {
			data, err := parseParasiteData(scanner.cfg, scanResult)
			if err != nil {
				logger.Println("error parsing parasite data:", err.Error())
			} else {
				// Have we processed this data already?
				if oldCounter, exists := scanner.lastCounter[data.Key]; exists && oldCounter == int(data.Counter) {
					return
				}
				scanner.lastCounter[data.Key] = int(data.Counter)
				scanner.channel <- data
			}
		}
	})

	if err != nil {
		panic("unable to start scanning: " + err.Error())
	}
}
