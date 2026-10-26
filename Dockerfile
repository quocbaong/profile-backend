# Build stage
FROM golang:alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api ./cmd/api

# Final production stage (~15MB image)
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/api .
COPY --from=builder /app/.env.example ./.env

EXPOSE 8080

CMD ["./api"]
