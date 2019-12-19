package main

// To run, edit config.yaml and execute:

// GOPATH=`pwd` go get github.com/eclipse/paho.mqtt.golang
// GOPATH=`pwd` go get github.com/goccy/go-yaml
// GOPATH=`pwd` go build hvac_ip_mqtt_bridge.go && ./hvac_ip_mqtt_bridge

// TODO(gsasha): docker
// TODO(gsasha): use go mod.
// TODO(gsasha): tests
// TODO(gsasha): export availability on mqtt
// TODO(gsasha): export health status on http

import (
	"flag"
	"github.com/gsasha/hvac_ip_mqtt_bridge/hvac/loader"
	"log"
	"net/http"
)

var configFile = flag.String("config_file", "config.yaml", "configuration file")

func main() {
	log.Printf("HVAC IP to MQTT Bridge starting up.")
	flag.Parse()
	devices, err := loader.Load(*configFile)
	if err != nil {
		log.Fatalf("Loading failed: %s", err)
	}
	log.Printf("Running configured devices")
	for _, device := range devices {
		device.Run()
	}
	log.Printf("Listening to HTTP port")
	http.ListenAndServe(":8080", nil)
	log.Printf("Done ListenAndServe")
}
