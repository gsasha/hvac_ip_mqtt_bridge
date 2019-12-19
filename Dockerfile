FROM golang:latest

LABEL maintainer="Sasha Gontmakher <gsasha@gmail.com>"

WORKDIR /data

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bridge

EXPOSE 80

CMD ["./bridge --config_file=/data/config.yaml"]

