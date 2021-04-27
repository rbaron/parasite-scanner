package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type MACAddr string

type MQTTParasiteConfig struct {
	SoilMoistureTopic   string `yaml:"soil_moisture_topic"`
	TemperatureTopic    string `yaml:"temperature_topic"`
	HumidityTopic       string `yaml:"humidity_topic"`
	BatteryVoltageTopic string `yaml:"battery_voltage_topic"`
	RSSITopic           string `yaml:"rssi_topic"`
}

type MQTTConfig struct {
	Host          string                          `yaml:"host"`
	Username      string                          `yaml:"username"`
	Password      string                          `yaml:"password"`
	ClientId      string                          `yaml:"client_id"`
	AutoDiscovery bool                            `yaml:"auto_discovery"`
	Registry      map[MACAddr]*MQTTParasiteConfig `yaml:"registry"`
}

type BLEConfig struct {
	MacOS struct {
		InferMACAddress  bool   `yaml:"infer_mac_address"`
		MACAddressPrefix string `yaml:"mac_address_prefix"`
	} `yaml:"macos"`
}

type Config struct {
	MQTT MQTTConfig `yaml:"mqtt"`
	BLE  BLEConfig
}

func ValidateMQTTParasiteConfig(cfg *MQTTParasiteConfig) error {
	if cfg.SoilMoistureTopic == "" {
		return fmt.Errorf("missing soil_moisture_topic")
	} else if cfg.TemperatureTopic == "" {
		return fmt.Errorf("missing temperature_topic")
	} else if cfg.HumidityTopic == "" {
		return fmt.Errorf("missing humidity_topic")
	} else if cfg.BatteryVoltageTopic == "" {
		return fmt.Errorf("missing battery_voltage_topic")
	}
	return nil
}

func ParseConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)

	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	for mac_addr, mqtt_cfg := range config.MQTT.Registry {
		if err := ValidateMQTTParasiteConfig(mqtt_cfg); err != nil {
			return nil, fmt.Errorf("%s: %s", mac_addr, err.Error())
		}
	}
	return config, nil
}
