FROM golang:1.23.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o alienos --tags libsecp256k1

FROM alpine:latest

COPY --from=builder /app/alienos /usr/local/bin/alienos

EXPOSE 7771

CMD ["alienos"]
