package hvac

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
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

	client     mqtt.Client
	controller Controller
}

func NewMQTT(broker string, clientId string, prefix string) *MQTT {
	log.Printf("Connecting to MQTT broker %s for %s, prefix %s", broker, clientId, prefix)
	m := &MQTT{
		clientId: clientId,
		prefix:   prefix,
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

func (m *MQTT) SetController(controller Controller) {
	m.controller = controller
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
	tokens := []mqtt.Token{
		m.client.Subscribe(m.prefix+"/"+powerCommandTopic, 0,
			func(client mqtt.Client, message mqtt.Message) {
				log.Println("Received", message.Topic(), string(message.Payload()))
				m.controller.SetPowerMode(string(message.Payload()))
			}),
		m.client.Subscribe(m.prefix+"/"+opModeCommandTopic, 0,
			func(client mqtt.Client, message mqtt.Message) {
				log.Println("Received", message.Topic(), string(message.Payload()))
				m.controller.SetOpMode(string(message.Payload()))
			}),
		m.client.Subscribe(m.prefix+"/"+fanModeCommandTopic, 0,
			func(client mqtt.Client, message mqtt.Message) {
				log.Println("Received", message.Topic(), string(message.Payload()))
				m.controller.SetFanMode(string(message.Payload()))
			}),
		m.client.Subscribe(m.prefix+"/"+temperatureCommandTopic, 0,
			func(client mqtt.Client, message mqtt.Message) {
				log.Println("Received", message.Topic(), string(message.Payload()))
				m.controller.SetTemperature(string(message.Payload()))
			}),
		// TODO(gsasha): subscribe to more commands.
	}
	for _, token := range tokens {
		if token.Wait() && token.Error() != nil {
			log.Printf("Error subscribing to topics: %s", m.clientId, token.Error())
			return
		}
	}
	log.Printf("Subscribed to topics for %s", m.clientId)
}

func (m *MQTT) UpdateAction(action string) {
	m.publish(actionTopic, action)
}
func (m *MQTT) UpdateOpMode(opMode string) {
	m.publish(opModeStateTopic, opMode)
}
func (m *MQTT) UpdateFanMode(fanMode string) {
	m.publish(fanModeStateTopic, fanMode)
}
func (m *MQTT) UpdateTemperature(temperature string) {
	m.publish(temperatureStateTopic, temperature)
}
func (m *MQTT) UpdateCurrentTemperature(temperature string) {
	m.publish(currentTemperatureStateTopic, temperature)
}
func (m *MQTT) UpdateAttributes(attributes map[string]string) {
	// TODO(gsasha): implement.
}
func (m *MQTT) publish(topic string, message string) {
	log.Println("mqtt publishing", m.prefix+"/"+topic, message)
	m.client.Publish(m.prefix+"/"+topic, 0, false, message)
}
