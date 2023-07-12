FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ENV TOKEN "TOKEN"
ENV GOD_ID "1063077630"

ENTRYPOINT ["go", "run", "main.go"]
