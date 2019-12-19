# hvac_ip_mqtt_bridge
A bridge to connect between ip-enabled HVAC units and mqtt (to be connected to HomeAssistant etc)
Currently supported models:

- Samsung 2878

## Sample config.yaml:
```yaml
mqtt:
  host: "10.10.10.10"
devices:
  - name: "my_ac"
    model: "samsungac2878"
    host: "10.10.10.20"
    mqtt_prefix: "hvac/my_ac"
    duid: "112233445566"
    auth_token: "11111111-2222-3333-4444-5555555555"
```
 
## Corresponding config entry in Home Assistant climate.yaml:
```yaml
- platform: mqtt
  name: "My Air Conditioner"
  unique_id: "climate.my_ac"
  power_command_topic: "hvac/my_ac/power/set"
  mode_state_topic: "hvac/my_ac/mode/state"
  mode_command_topic: "hvac/my_ac/mode/set"
  action_topic: "hvac/my_ac/action"
  fan_mode_state_topic: "hvac/my_ac/fan_mode/state"
  fan_mode_command_topic: "test/samsunagc_mqtt/fan_mode/set"
  temperature_state_topic: "hvac/my_ac/temperature/state"
  temperature_command_topic: "hvac/my_ac/temperature/set"
  current_temperature_topic: "hvac/my_ac/current_temperature/state"
  precision: 0.1
  retain: false
  initial: 23
  device:
    manufacturer: Samsung
    identifiers: ["samsungac_my_ac"]

```
