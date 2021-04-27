package main

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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

// type AutoDiscoveryMsg struct {
// 	UnitOfMeasument   string `json:"unit_of_measurement"`
// 	Name              string `json:"name"`
// 	StateTopic        string `json:"state_topic"`
// 	AvailabilityTopic string `json:"availability_topic"`
// 	UniqueID          string `json:"unique_id"`
// 	// Device            struct {
// 	// 	Identifier string `json:"identifier"`
// 	// 	Name       string `json:"name"`
// 	// } `json:"device"`
// }

// func MakeAutoDiscoveryMessages() {

// }

func (client *MQTTClient) Publish(topic string, msg string, retained bool, qos byte) mqtt.Token {
	logger.Printf("[mqtt] Publishing %s to %s\n", msg, topic)
	return client.client.Publish(topic, qos, retained, msg)
}

func (client *MQTTClient) Run() {
	if token := client.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// TODO(rbaron): maybe publish auto-discovery data.

	for msg := range client.outgoing {
		deviceConfig, exists := client.config.Registry[MACAddr(msg.Key)]
		if !exists {
			logger.Printf("Received valid BLE broadcast from %s, but it's not configured for MQTT\n", msg.Key)
			continue
		}
		client.Publish(deviceConfig.SoilMoistureTopic, fmt.Sprintf("%.1f", msg.SoilMoisture), false, 1)
		client.Publish(deviceConfig.SoilMoistureTopic, fmt.Sprintf("%.1f", msg.SoilMoisture), false, 1)
		client.Publish(deviceConfig.TemperatureTopic, fmt.Sprintf("%.1f", msg.TempCelcius), false, 1)
		client.Publish(deviceConfig.HumidityTopic, fmt.Sprintf("%.1f", msg.Humidity), false, 1)
		client.Publish(deviceConfig.BatteryVoltageTopic, fmt.Sprintf("%.1f", msg.BatteryVoltage), false, 1)
		client.Publish(deviceConfig.RSSITopic, fmt.Sprintf("%d", msg.RSSI), false, 1)
	}
}
