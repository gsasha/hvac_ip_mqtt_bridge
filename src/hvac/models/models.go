package models

import (
	"fmt"
	"hvac"
	"hvac/models/samsung"
)

func NewController(
	model string, name string,
	host, port, duid, authToken string) (hvac.Controller, error) {
	switch model {
	case "samsungac2878":
		return samsung.NewSamsungAC2878(name, host, port, duid, authToken), nil
	}
	return nil, fmt.Errorf("Model not supported: %s", model)
}
