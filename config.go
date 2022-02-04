package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type MACAddr string

type ParasiteConfig struct {
	Name string `yaml:"name"`
}

const kBaseMQTTTopic string = "parasite-scanner/sensor/%s_%s/state"

func (cfg *ParasiteConfig) NormalizedName() string {
	return strings.Replace(strings.ToLower(cfg.Name), " ", "_", -1)
}

func (cfg *ParasiteConfig) TemperatureTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "temperature")
}

func (cfg *ParasiteConfig) HumidityTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "humidity")
}

func (cfg *ParasiteConfig) BatteryPercentageTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "battery_percentage")
}

func (cfg *ParasiteConfig) BatteryVoltageTopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "battery_voltage")
}

func (cfg *ParasiteConfig) RSSITopic() string {
	return fmt.Sprintf(kBaseMQTTTopic, cfg.NormalizedName(), "rssi")
}

type DEVICEConfig struct {
	Registry map[MACAddr]*ParasiteConfig `yaml:"registry"`
}

type PROMConfig struct {
	Enable bool   `yaml:"enable"`
	Host   string `yaml:"host"`
}

type MQTTConfig struct {
	Enable        bool   `yaml:"enable"`
	Host          string `yaml:"host"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	ClientId      string `yaml:"client_id"`
	AutoDiscovery bool   `yaml:"auto_discovery"`
}

type BLEConfig struct {
	VendorPrefix string `yaml:"vendor_prefix"`
	MacOS        struct {
		InferMACAddress  bool   `yaml:"infer_mac_address"`
		MACAddressPrefix string `yaml:"mac_address_prefix"`
	} `yaml:"macos"`
}

type Config struct {
	PROM   PROMConfig   `yaml:"prometheus"`
	MQTT   MQTTConfig   `yaml:"mqtt"`
	DEVICE DEVICEConfig `yaml:"device"`
	BLE    BLEConfig    `yaml:"ble"`
}

func ValidateParasiteConfig(cfg *ParasiteConfig) error {
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

	for macAddr, globalCfg := range config.DEVICE.Registry {
		if err := ValidateParasiteConfig(globalCfg); err != nil {
			return nil, fmt.Errorf("%s: %s", macAddr, err.Error())
		}
		// Normalize MAC address (to lowercase).
		delete(config.DEVICE.Registry, macAddr)
		normalizedMACAddr := strings.ToLower(string(macAddr))
		config.DEVICE.Registry[MACAddr(normalizedMACAddr)] = globalCfg
	}
	return config, nil
}
