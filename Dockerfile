FROM golang:tip-alpine3.23 AS builder
WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/node ./cmd/node

FROM alpine:latest

WORKDIR /app
COPY --from=builder /bin/node .

CMD ["./node"]