package main

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient implements the DataSubscriber interface.
// It will prepate and publish incoming data to their respective MQTT topics.
type MQTTClient struct {
	client   mqtt.Client
	outgoing chan *ParasiteData
	config   *MQTTConfig
}

func MakeMQTTClient(cfg *MQTTConfig) *MQTTClient {
	opts := mqtt.
		NewClientOptions().
		AddBroker(cfg.Host).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetClientID(cfg.ClientId)

	// opts.SetKeepAlive(1 * time.Second)
	// opts.SetPingTimeout(1 * time.Second)

	client := mqtt.NewClient(opts)
	return &MQTTClient{
		client:   client,
		outgoing: make(chan *ParasiteData),
		config:   cfg,
	}
}

type AutoDiscoveryDeviceInfo struct {
	Identifiers  string `json:"identifiers"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
}

type AutoDiscoveryPayload struct {
	DeviceClass       string                   `json:"device_class"`
	UnitOfMeasument   string                   `json:"unit_of_measurement"`
	Name              string                   `json:"name"`
	StateTopic        string                   `json:"state_topic"`
	AvailabilityTopic string                   `json:"availability_topic"`
	UniqueID          string                   `json:"unique_id"`
	Device            *AutoDiscoveryDeviceInfo `json:"device"`
}

type AutoDiscoveryMsg struct {
	Topic   string
	Payload AutoDiscoveryPayload
}

func makeAutoDiscoveryMessages(deviceConfig *MQTTParasiteConfig) []*AutoDiscoveryMsg {
	device := &AutoDiscoveryDeviceInfo{
		Identifiers:  "parasite-scanner",
		Name:         "parasite-scanner",
		Manufacturer: "rbaron",
	}
	return []*AutoDiscoveryMsg{
		{Topic: fmt.Sprintf("homeassistant/sensor/parasite-scanner/%s_soil_moisture/config", deviceConfig.NormalizedName()),
			Payload: AutoDiscoveryPayload{
				DeviceClass:       "humidity",
				UnitOfMeasument:   "%",
				Name:              fmt.Sprintf("%s Soil Moisture", deviceConfig.Name),
				StateTopic:        deviceConfig.SoilMoistureTopic(),
				UniqueID:          fmt.Sprintf("%s_soil_moisture", deviceConfig.NormalizedName()),
				AvailabilityTopic: "parasite-scanner/status",
				Device:            device,
			}},
		{Topic: fmt.Sprintf("homeassistant/sensor/parasite-scanner/%s_temperature/config", deviceConfig.NormalizedName()),
			Payload: AutoDiscoveryPayload{
				DeviceClass:       "temperature",
				UnitOfMeasument:   "°C",
				Name:              fmt.Sprintf("%s Temperature", deviceConfig.Name),
				StateTopic:        deviceConfig.TemperatureTopic(),
				UniqueID:          fmt.Sprintf("%s_temperature", deviceConfig.NormalizedName()),
				AvailabilityTopic: "parasite-scanner/status",
				Device:            device,
			}},
		{Topic: fmt.Sprintf("homeassistant/sensor/parasite-scanner/%s_humidity/config", deviceConfig.NormalizedName()),
			Payload: AutoDiscoveryPayload{
				DeviceClass:       "humidity",
				UnitOfMeasument:   "%",
				Name:              fmt.Sprintf("%s Humidity", deviceConfig.Name),
				StateTopic:        deviceConfig.HumidityTopic(),
				UniqueID:          fmt.Sprintf("%s_humidity", deviceConfig.NormalizedName()),
				AvailabilityTopic: "parasite-scanner/status",
				Device:            device,
			}},
		{Topic: fmt.Sprintf("homeassistant/sensor/parasite-scanner/%s_battery_voltage/config", deviceConfig.NormalizedName()),
			Payload: AutoDiscoveryPayload{
				DeviceClass:       "voltage",
				UnitOfMeasument:   "V",
				Name:              fmt.Sprintf("%s Battery Voltage", deviceConfig.Name),
				StateTopic:        deviceConfig.BatteryVoltageTopic(),
				UniqueID:          fmt.Sprintf("%s_battery_voltage", deviceConfig.NormalizedName()),
				AvailabilityTopic: "parasite-scanner/status",
				Device:            device,
			}},
		{Topic: fmt.Sprintf("homeassistant/sensor/parasite-scanner/%s_rssi/config", deviceConfig.NormalizedName()),
			Payload: AutoDiscoveryPayload{
				DeviceClass:       "signal_strength",
				UnitOfMeasument:   "dB",
				Name:              fmt.Sprintf("%s RSSI", deviceConfig.Name),
				StateTopic:        deviceConfig.RSSITopic(),
				UniqueID:          fmt.Sprintf("%s_rssi", deviceConfig.NormalizedName()),
				AvailabilityTopic: "parasite-scanner/status",
				Device:            device,
			}},
	}
}

func (client *MQTTClient) publishData(deviceConfig *MQTTParasiteConfig, data *ParasiteData) {
	client.Publish(deviceConfig.SoilMoistureTopic(), fmt.Sprintf("%.1f", data.SoilMoisture), false, 1)
	client.Publish(deviceConfig.TemperatureTopic(), fmt.Sprintf("%.1f", data.TempCelcius), false, 1)
	client.Publish(deviceConfig.HumidityTopic(), fmt.Sprintf("%.1f", data.Humidity), false, 1)
	client.Publish(deviceConfig.BatteryVoltageTopic(), fmt.Sprintf("%.1f", data.BatteryVoltage), false, 1)
	client.Publish(deviceConfig.RSSITopic(), fmt.Sprintf("%d", data.RSSI), false, 1)
}

func (client *MQTTClient) Publish(topic string, msg string, retained bool, qos byte) mqtt.Token {
	logger.Printf("[mqtt] Publishing %s to %s\n", msg, topic)
	return client.client.Publish(topic, qos, retained, msg)
}

func (client *MQTTClient) Ingest(data *ParasiteData) {
	client.outgoing <- data
}

func (client *MQTTClient) Run() {
	if token := client.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if client.config.AutoDiscovery {
		for macAddr, deviceConfig := range client.config.Registry {
			logger.Printf("Generating auto-discovery messages for %s\n", macAddr)
			for _, msg := range makeAutoDiscoveryMessages(deviceConfig) {
				payload, _ := json.Marshal(msg.Payload)
				client.Publish(msg.Topic, string(payload), true, 1)
			}
		}
	}

	for data := range client.outgoing {
		deviceConfig, exists := client.config.Registry[MACAddr(data.Key)]
		if !exists {
			logger.Printf("Received valid BLE broadcast from %s, but it's not configured for MQTT\n", data.Key)
			continue
		}
		client.publishData(deviceConfig, data)
	}
}
