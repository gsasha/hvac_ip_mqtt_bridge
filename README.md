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
 
