package samsung

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/gsasha/hvac_ip_mqtt_bridge/hvac/base"
	"log"
	"strings"
	"text/template"
	"time"
)

type SamsungAC2878 struct {
	name      string
	host      string
	port      string
	authToken string
	duid      string

	connection    base.Connection
	stateNotifier base.StateNotifier

	online             bool
	err                string
	powerMode          string
	opMode             string
	fanMode            string
	temperature        string
	currentTemperature string
	purifyMode	   string
	swingMode	   string
	attrs              map[string]string
}

func NewSamsungAC2878(name string, host, port, duid, authToken string) *SamsungAC2878 {
	if port == "" {
		port = "2878"
	}
	return &SamsungAC2878{
		name:       name,
		host:       host,
		port:       port,
		authToken:  authToken,
		duid:       duid,
		connection: base.NewTLSSocketConnection(),
		attrs:      make(map[string]string),
	}
}

func (c *SamsungAC2878) SetStateNotifier(stateNotifier base.StateNotifier) {
	c.stateNotifier = stateNotifier
}

func (c *SamsungAC2878) Connect() {
	c.connection.Connect(c.host, c.port, c)
	go func() {
		for range time.Tick(time.Second * 60) {
			c.sendDeviceStateRequest()
		}
	}()
}

var (
	authenticateTemplate = template.Must(template.New("authenticate").Parse(
		`<Request Type="AuthToken"><User Token="{{.token}}" /></Request>
`))
	deviceStateTemplate = template.Must(template.New("deviceState").Parse(
		`<Request Type="DeviceState" DUID="{{.duid}}"></Request>
`))
	setPowerModeTemplate = template.Must(template.New("setPowerMode").Parse(
		`<Request Type="DeviceControl"><Control CommandID="AC_FUN_POWER" DUID="{{.duid}}"><Attr ID="AC_FUN_POWER" Value="{{.value}}" /></Control></Request>
`))
	setModeTemplate = template.Must(template.New("setMode").Parse(
		`<Request Type="DeviceControl"><Control CommandID="AC_FUN_OPMODE" DUID="{{.duid}}"><Attr ID="AC_FUN_OPMODE" Value="{{.value}}" /></Control></Request>
`))
	setFanModeTemplate = template.Must(template.New("setFanMode").Parse(
		`<Request Type="DeviceControl"><Control CommandID="AC_FUN_WINDLEVEL" DUID="{{.duid}}"><Attr ID="AC_FUN_WINDLEVEL" Value="{{.value}}" /></Control></Request>
`))
	setTemperatureTemplate = template.Must(template.New("setTemperature").Parse(
		`<Request Type="DeviceControl"><Control CommandID="AC_FUN_TEMPSET" DUID="{{.duid}}"><Attr ID="AC_FUN_TEMPSET" Value="{{.value}}" /></Control></Request>
`))
	setPurifyModeTemplate = template.Must(template.New("setPurifyMode").Parse(
		`<Request Type="DeviceControl"><Control CommandID="AC_ADD_SPI" DUID="{{.duid}}"><Attr ID="AC_ADD_SPI" Value="{{.value}}" /></Control></Request>
`))
        setSwingModeTemplate = template.Must(template.New("setSwingMode").Parse(
		`<Request Type="DeviceControl"><Control CommandID="AC_FUN_DIRECTION" DUID="{{.duid}}"><Attr ID="AC_FUN_DIRECTION" Value="{{.value}}" /></Control></Request>
`))

)

func (c *SamsungAC2878) SetPowerMode(powerMode string) {
	c.sendMessage(setPowerModeTemplate, map[string]string{
		"value": PowerModeToAC(powerMode),
		"duid":  c.duid,
	})
}

func (c *SamsungAC2878) SetOpMode(mode string) {
	if mode == "off" {
		c.sendMessage(setPowerModeTemplate, map[string]string{
			"value": "Off",
			"duid":  c.duid,
		})
	} else {
		c.sendMessage(setModeTemplate, map[string]string{
			"value": OpModeToAC(mode),
			"duid":  c.duid,
		})
	}
}

func (c *SamsungAC2878) SetFanMode(fanMode string) {
	c.sendMessage(setFanModeTemplate, map[string]string{
		"value": FanModeToAC(fanMode),
		"duid":  c.duid,
	})
}

func (c *SamsungAC2878) SetTemperature(temperature string) {
	c.sendMessage(setTemperatureTemplate, map[string]string{
		"value": temperature,
		"duid":  c.duid,
	})
}

func (c *SamsungAC2878) SetPurifyMode(purifyMode string) {
	c.sendMessage(setPurifyModeTemplate, map[string]string{
		"value": purifyMode,
		"duid":  c.duid,
	})
}
func (c *SamsungAC2878) SetSwingMode(swingMode string) {
	c.sendMessage(setSwingModeTemplate, map[string]string{
		"value": swingMode,
		"duid":  c.duid,
	})
}

type Response struct {
	XMLName     xml.Name `xml:"Response"`
	Type        string   `xml:"Type,attr"`
	Status      string   `xml:"Status,attr"`
	DeviceState DeviceState
	Inner       []byte `xml:",innerxml"`
}

type Update struct {
	XMLName xml.Name `xml:"Update"`
	Type    string   `xml:"Type,attr"`
	Status  Status
}
type Attr struct {
	XMLName xml.Name `xml:"Attr"`
	ID      string   `xml:"ID,attr"`
	Type    string   `xml:"Type,attr"`
	Value   string   `xml:"Value,attr"`
}
type Status struct {
	XMLName xml.Name `xml:"Status"`
	DUID    string   `xml:"DUID"`
	GroupID string   `xml:GroupID,attr"`
	ModelID string   `xml:ModelID,attr"`
	Attr    []Attr
}
type Device struct {
	XMLName xml.Name `xml:"Device"`
	DUID    string   `xml:"DUID,attr"`
	GroupID string   `xml:"GroupID,attr"`
	ModelID string   `xml:"ModelID,attr"`
	Attr    []Attr
}
type DeviceState struct {
	XMLName xml.Name `xml:"DeviceState"`
	Device  Device
}

func (c *SamsungAC2878) OnConnectionEstablished() {
	log.Printf("Established connection to %s", c.name)
	c.connection.ExpectRead()
}

func (c *SamsungAC2878) HandleMessage(message []byte) {
	log.Printf("Received message from %s: %s", c.name, string(message))

	if string(message) == "DPLUG-1.6\n" {
		log.Printf("Connection hello received from %s", c.name)
		c.connection.ExpectRead()
	}
	var update Update
	if err := xml.Unmarshal(message, &update); err == nil {
		c.handleUpdate(&update)
		return
	}
	var response Response
	if err := xml.Unmarshal(message, &response); err == nil {
		c.handleResponse(&response)
		return
	}
}

func (c *SamsungAC2878) handleUpdate(update *Update) error {
	switch update.Type {
	case "InvalidateAccount":
		c.handleInvalidateAccount()
	case "Status":
		c.handleUpdateStatus(&update.Status)
	default:
		log.Println("Error: %s unknown update type", c.name, update.Type)
		return nil
	}
	return nil
}

func (c *SamsungAC2878) handleResponse(response *Response) error {
	switch response.Type {
	case "AuthToken":
		c.handleAuthToken(response.Status)
	case "DeviceState":
		c.handleDeviceState(&response.DeviceState)
	case "DeviceControl":
		c.handleDeviceControl(response.Status)
	default:
		fmt.Println("Error: %s got unknown response", c.name, response.Type)
	}
	return nil
}

func (c *SamsungAC2878) handleInvalidateAccount() {
	c.sendMessage(authenticateTemplate, map[string]string{
		"duid":  c.duid,
		"token": c.authToken,
	})
}

func (c *SamsungAC2878) sendDeviceStateRequest() {
	c.sendMessage(deviceStateTemplate, map[string]string{
		"duid": c.duid,
	})
}

func (c *SamsungAC2878) handleAuthToken(status string) {
	if status == "Okay" {
		c.online = true
	} else {
		c.online = false
	}
	c.sendDeviceStateRequest()
}

func (c *SamsungAC2878) handleDeviceControl(status string) {
	if status == "Okay" {
		c.err = ""
	} else {
		c.err = status
	}
}

func (c *SamsungAC2878) handleUpdateStatus(status *Status) {
	if status == nil {
		fmt.Println("Error: No status")
		return
	}
	c.handleAttributes(status.Attr)
	c.notifyState()
}

func (c *SamsungAC2878) handleDeviceState(deviceState *DeviceState) {
	c.handleAttributes(deviceState.Device.Attr)
	c.notifyState()
}

func (c *SamsungAC2878) notifyState() {
	if c.stateNotifier == nil {
		fmt.Println("Error: want to notify state, but no notifer defined")
		return
	}
	if strings.ToLower(c.powerMode) == "off" {
		c.stateNotifier.UpdateOpMode(OpModeFromAC("Off"))
	} else {
		c.stateNotifier.UpdateOpMode(OpModeFromAC(c.opMode))
	}
	c.stateNotifier.UpdateFanMode(FanModeFromAC(c.fanMode))
	c.stateNotifier.UpdateTemperature(c.temperature)
	c.stateNotifier.UpdateCurrentTemperature(c.currentTemperature)
	c.stateNotifier.UpdatePurifyMode(c.purifyMode)
	c.stateNotifier.UpdateSwingMode(c.swingMode)
	c.stateNotifier.UpdateAttributes(c.attrs)
}

func (c *SamsungAC2878) handleAttributes(attrs []Attr) {
	for _, attr := range attrs {
		c.attrs[attr.Type] = attr.Value
		switch attr.ID {
		case "AC_FUN_POWER":
			c.powerMode = attr.Value
		case "AC_FUN_OPMODE":
			c.opMode = attr.Value
		case "AC_FUN_TEMPSET":
			c.temperature = attr.Value
		case "AC_FUN_TEMPNOW":
			c.currentTemperature = attr.Value
		case "AC_FUN_WINDLEVEL":
			c.fanMode = attr.Value
		case "AC_ADD_SPI":
			c.purifyMode = attr.Value
		case "AC_FUN_DIRECTION":
			c.swingMode = attr.Value
		}
	}
}

func (c *SamsungAC2878) sendMessage(messageTemplate *template.Template, data map[string]string) {
	var buf bytes.Buffer
	messageTemplate.Execute(&buf, data)
	log.Printf("sending request to %s [%s]\n", c.name, buf.String())
	c.connection.SendMessage(buf.Bytes())
}
