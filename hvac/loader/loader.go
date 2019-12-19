package loader

import (
	"fmt"
	yaml "github.com/goccy/go-yaml"
	"github.com/gsasha/hvac_ip_mqtt_bridge/hvac/base"
	"github.com/gsasha/hvac_ip_mqtt_bridge/hvac/models"
	"io/ioutil"
)

type Config struct {
	MQTT    *MQTTConfig    `yaml:"mqtt"`
	Devices []DeviceConfig `yaml:"devices"`
}

type MQTTConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Protocol string `yaml:"protocol"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DeviceConfig struct {
	Name       string `yaml:"name"`
	Model      string `yaml:"model"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	MQTTPrefix string `yaml:"mqtt_prefix"`
	DUID       string `yaml:"duid"`
	AuthToken  string `yaml:"auth_token"`
}

type Device struct {
	mqtt       *base.MQTT
	controller base.Controller
}

func NewDevice(mqttConfig MQTTConfig, deviceConfig DeviceConfig) (*Device, error) {
	protocol := mqttConfig.Protocol
	if protocol == "" {
		protocol = "tcp"
	}
	host := mqttConfig.Host
	if host == "" {
		return nil, fmt.Errorf("MQTT host not given for %s", deviceConfig.Name)
	}
	port := mqttConfig.Port
	if port == "" {
		port = "1883"
	}
	mqttBroker := fmt.Sprintf("%s://%s:%s", protocol, mqttConfig.Host, port)
	mqttClientId := fmt.Sprintf("hvac_ip_mqtt_bridge_%s", deviceConfig.Name)
	mqtt := base.NewMQTT(mqttBroker, mqttClientId)
	controller, err := models.NewController(
		deviceConfig.Model,
		deviceConfig.Name,
		deviceConfig.Host,
		deviceConfig.Port,
		deviceConfig.DUID,
		deviceConfig.AuthToken)

	notifier := mqtt.RegisterController(deviceConfig.Name, deviceConfig.MQTTPrefix, controller)
	controller.SetStateNotifier(notifier)

	if err != nil {
		return nil, err
	}
	return &Device{
		mqtt:       mqtt,
		controller: controller,
	}, nil
}

func (device *Device) Run() {
	device.controller.Connect()
	device.mqtt.Connect()
}

func Load(configFile string) ([]*Device, error) {
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, err
	}
	if config.MQTT == nil {
		return nil, fmt.Errorf("mqtt missing in configuration")
	}
	var devices []*Device
	for _, deviceConfig := range config.Devices {
		device, err := NewDevice(*config.MQTT, deviceConfig)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}
