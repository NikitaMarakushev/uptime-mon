FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /uptime-monitor main.go

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /uptime-monitor /app/uptime-monitor

USER nonroot:nonroot

EXPOSE 9116

ENTRYPOINT ["/app/uptime-monitor"]