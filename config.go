package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type MACAddr string

type MQTTParasiteConfig struct {
	Name string `yaml:"name"`
}

const kBaseMQTTTopic string = "parasite-scanner/sensor/%s_%s/state"

func (cfg *MQTTParasiteConfig) NormalizedName() string {
	return strings.Replace(strings.ToLower(cfg.Name), " ", "_", -1)
}

func (cfg *MQTTParasiteConfig) SoilMoistureTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "soil_moisture")
}

func (cfg *MQTTParasiteConfig) TemperatureTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "temperature")
}

func (cfg *MQTTParasiteConfig) HumidityTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "humidity")
}

func (cfg *MQTTParasiteConfig) BatteryVoltageTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "battery_voltage")
}

func (cfg *MQTTParasiteConfig) RSSITopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "rssi")
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
	if cfg.Name == "" {
		return fmt.Errorf("missing name")
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

	for macAddr, mqttCfg := range config.MQTT.Registry {
		if err := ValidateMQTTParasiteConfig(mqttCfg); err != nil {
			return nil, fmt.Errorf("%s: %s", macAddr, err.Error())
		}
		// Normalize MAC address (to lowercase).
		delete(config.MQTT.Registry, macAddr)
		normalizedMACAddr := strings.ToLower(string(macAddr))
		config.MQTT.Registry[MACAddr(normalizedMACAddr)] = mqttCfg
	}
	return config, nil
}
