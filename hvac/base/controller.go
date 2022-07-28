package base

type StateNotifier interface {
	UpdateAction(action string)
	UpdateOpMode(mode string)
	UpdateFanMode(fanMode string)
	UpdateTemperature(temperature string)
	UpdateCurrentTemperature(temperature string)
	UpdatePurifyMode(purifyMode string)
	UpdateAttributes(attributes map[string]string)
}

type Controller interface {
	SetStateNotifier(stateNotifier StateNotifier)
	Connect()
	SetPowerMode(powerMode string)
	SetOpMode(mode string)
	SetFanMode(fanMode string)
	SetTemperature(temperature string)
	SetPurifyMode(purifyMode string)
}
