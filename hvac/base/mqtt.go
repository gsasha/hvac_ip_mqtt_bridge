package base

import (
	"fmt"
	"log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	availabilityTopic            = "mode/availability"
	powerCommandTopic            = "power/set"
	opModeCommandTopic           = "mode/set"
	opModeStateTopic             = "mode/state"
	actionTopic                  = "action"
	currentTemperatureStateTopic = "current_temperature/state"
	temperatureCommandTopic      = "temperature/set"
	temperatureStateTopic        = "temperature/state"
	fanModeCommandTopic          = "fan_mode/set"
	fanModeStateTopic            = "fan_mode/state"
)

type MQTT struct {
	clientId string
	prefix   string

	client      mqtt.Client
	controllers map[string]Controller
	prefixes    map[string]string
}

type MQTTNotifier struct {
	mqtt   *MQTT
	prefix string
}

func (m *MQTTNotifier) UpdateAction(action string) {
	m.mqtt.updateAction(m.prefix, action)
}
func (m *MQTTNotifier) UpdateOpMode(opMode string) {
	m.mqtt.updateOpMode(m.prefix, opMode)
}
func (m *MQTTNotifier) UpdateFanMode(fanMode string) {
	m.mqtt.updateFanMode(m.prefix, fanMode)
}
func (m *MQTTNotifier) UpdateTemperature(temperature string) {
	m.mqtt.updateTemperature(m.prefix, temperature)
}
func (m *MQTTNotifier) UpdateCurrentTemperature(temperature string) {
	m.mqtt.updateCurrentTemperature(m.prefix, temperature)
}
func (m *MQTTNotifier) UpdateAttributes(attributes map[string]string) {
	m.mqtt.updateAttributes(m.prefix, attributes)
}

func NewMQTT(broker string, clientId string) *MQTT {
	log.Printf("Connecting to MQTT broker %s for %s", broker, clientId)
	m := &MQTT{
		clientId:    clientId,
		controllers: make(map[string]Controller),
		prefixes:    make(map[string]string),
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID("samsungac_mqtt")
	options.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("Connection established to %s:%s", clientId, broker)
		m.subscribeTopics()
	})
	options.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost to %s:%s %s", clientId, broker, err)
	})
	options.SetAutoReconnect(true)

	m.client = mqtt.NewClient(options)
	return m
}

func (m *MQTT) RegisterController(id string, prefix string, controller Controller) StateNotifier {
	m.controllers[id] = controller
	m.prefixes[id] = prefix
	return &MQTTNotifier{
		mqtt:   m,
		prefix: prefix,
	}
}

func (m *MQTT) Connect() {
	token := m.client.Connect()
	if token.Wait() && token.Error() == nil {
		log.Println("MQTT Connection succeeded:", m.client.IsConnectionOpen())
	} else {
		log.Println("MQTT Connection failed:", token.Error())
	}
}

func (m *MQTT) subscribeTopics() {
	for controllerId := range m.controllers {
		prefix := m.prefixes[controllerId]
		key := fmt.Sprintf("%s", controllerId)
		log.Printf("subscribing to prefix %s for %s", prefix, controllerId)
		tokens := []mqtt.Token{
			m.client.Subscribe(prefix+"/"+powerCommandTopic, 0,
				func(client mqtt.Client, message mqtt.Message) {
					log.Printf("Received %s:%s:%s", key, message.Topic(), string(message.Payload()))
					m.controllers[key].SetPowerMode(string(message.Payload()))
				}),
			m.client.Subscribe(prefix+"/"+opModeCommandTopic, 0,
				func(client mqtt.Client, message mqtt.Message) {
					log.Println("Received %s:%s:%s", key, message.Topic(), string(message.Payload()))
					m.controllers[key].SetOpMode(string(message.Payload()))
				}),
			m.client.Subscribe(prefix+"/"+fanModeCommandTopic, 0,
				func(client mqtt.Client, message mqtt.Message) {
					log.Println("Received %s:%s:%s", key, message.Topic(), string(message.Payload()))
					m.controllers[key].SetFanMode(string(message.Payload()))
				}),
			m.client.Subscribe(prefix+"/"+temperatureCommandTopic, 0,
				func(client mqtt.Client, message mqtt.Message) {
					log.Println("Received %s:%s:%s", key, message.Topic(), string(message.Payload()))
					m.controllers[key].SetTemperature(string(message.Payload()))
				}),
			// TODO(gsasha): subscribe to more commands.
		}
		for _, token := range tokens {
			if token.Wait() && token.Error() != nil {
				log.Printf("Error subscribing to topics %s: %s", controllerId, token.Error())
				return
			}
		}
		log.Printf("Subscribed to topics for %s", controllerId)
	}
}

func (m *MQTT) updateAction(prefix string, action string) {
	m.publish(prefix, actionTopic, action)
}
func (m *MQTT) updateOpMode(prefix string, opMode string) {
	m.publish(prefix, opModeStateTopic, opMode)
}
func (m *MQTT) updateFanMode(prefix string, fanMode string) {
	m.publish(prefix, fanModeStateTopic, fanMode)
}
func (m *MQTT) updateTemperature(prefix string, temperature string) {
	m.publish(prefix, temperatureStateTopic, temperature)
}
func (m *MQTT) updateCurrentTemperature(prefix string, temperature string) {
	m.publish(prefix, currentTemperatureStateTopic, temperature)
}
func (m *MQTT) updateAttributes(prefix string, attributes map[string]string) {
	// TODO(gsasha): implement.
}
func (m *MQTT) publish(prefix string, topic string, message string) {
	log.Println("mqtt publishing", prefix+"/"+topic, message)
	m.client.Publish(prefix+"/"+topic, 0, false, message)
}
