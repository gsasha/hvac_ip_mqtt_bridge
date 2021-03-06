# docker build --pull --label gsasha/hvac_ip_mqtt_bridge:latest -t gsasha/hvac_ip_mqtt_bridge:latest .
# docker build --pull --label gsasha/hvac_ip_mqtt_bridge:latest_arm -t gsasha/hvac_ip_mqtt_bridge:latest_arm .
# docker build --label gsasha/hvac_ip_mqtt_bridge:latest .
# docker push gsasha/hvac_ip_mqtt_bridge:latest
# docker push gsasha/hvac_ip_mqtt_bridge:latest_arm
FROM golang:latest

LABEL maintainer="Sasha Gontmakher <gsasha@gmail.com>"

WORKDIR /data

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bridge

EXPOSE 8080

CMD ["./bridge", "--config_file=/config/config.yaml"]

