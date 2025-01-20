package main

import (
	"encoding/json"
	"os"
	"log"
)

// Config structure for settings
type Config struct {
	Co2Sensor struct {
		IP  string `json:"ip"`
	} `json:"co2Sensor"`
	Switchbot struct {
		MAC string `json:"mac"`
	} `json:"switchbot"`
}

// LoadConfig reads the configuration file
func LoadConfig(filePath string) Config {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}
	return config
}
