FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wallet-service ./

FROM alpine:3.18
WORKDIR /app

COPY --from=builder /src/wallet-service ./wallet-service
COPY config.env ./config.env

EXPOSE 8080

CMD ["./wallet-service"]