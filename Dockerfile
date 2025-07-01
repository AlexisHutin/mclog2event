# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /mclog2event .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /mclog2event /mclog2event
ENTRYPOINT ["/mclog2event"]
