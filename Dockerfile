FROM golang:1.23-alpine3.21

WORKDIR /subpub-service

COPY . .

RUN go mod tidy

WORKDIR /subpub-service/cmd/app

VOLUME /subpub-service/logs

EXPOSE 9736

ENTRYPOINT [ "go", "run", "main.go"]
