FROM golang:1.21-bookworm AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o nucleus-api .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget git procps

COPY --from=builder /app/nucleus-api /usr/local/bin/nucleus-api

EXPOSE 8080

CMD ["nucleus-api"]
