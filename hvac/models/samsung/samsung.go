package samsung

import (
	"strings"
)

type translationEntry struct {
	mqtt string
	ac   string
}

func toAc(value string, table []translationEntry) string {
	for _, e := range table {
		if strings.ToLower(value) == strings.ToLower(e.mqtt) {
			return e.ac
		}
	}
	return value
}

func fromAc(value string, table []translationEntry) string {
	for _, e := range table {
		if strings.ToLower(value) == strings.ToLower(e.ac) {
			return e.mqtt
		}
	}
	return strings.ToLower(value)
}

var powerModeTable = []translationEntry{
	{"ON", "On"},
	{"off", "Off"},
}

func PowerModeToAC(mode string) string   { return toAc(mode, powerModeTable) }
func PowerModeFromAC(mode string) string { return fromAc(mode, powerModeTable) }

var opModeTable = []translationEntry{
	{"cool", "Cool"},
	{"heat", "Heat"},
	{"dry", "Dry"},
	{"auto", "Auto"},
	{"fan_only", "Wind"},
	{"off", "Off"},
}

func OpModeToAC(mode string) string   { return toAc(mode, opModeTable) }
func OpModeFromAC(mode string) string { return fromAc(mode, opModeTable) }

var fanModeTable = []translationEntry{
	{"auto", "Auto"},
	{"low", "Low"},
	{"medium", "Mid"},
	{"high", "Turbo"},
}

func FanModeToAC(mode string) string   { return toAc(mode, fanModeTable) }
func FanModeFromAC(mode string) string { return fromAc(mode, fanModeTable) }
