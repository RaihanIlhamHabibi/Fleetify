FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY backend/go.mod ./
RUN go mod download
COPY backend/ ./
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /fleetify .
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /fleetify .
COPY frontend/ ./frontend/
RUN mkdir -p /app/uploads
EXPOSE 8080
CMD ["./fleetify"]
