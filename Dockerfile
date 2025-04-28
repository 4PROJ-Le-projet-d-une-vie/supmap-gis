FROM golang:1.24 AS builder

WORKDIR /app

# Install dependencies and swaggo
COPY go.mod go.sum ./
RUN go mod tidy
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy whole sources
COPY . .

# Build application
RUN swag init --parseDependency --parseInternal --dir ./cmd/api/,./internal/ --outputTypes go,yaml
RUN CGO_ENABLED=0 GOOS=linux go build -o supmap-gis cmd/api/main.go

# Build final image
FROM golang:1.24-alpine
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/supmap-gis .
COPY --from=builder /app/docs ./docs

# Default values
ENV API_SERVER_HOST=0.0.0.0
ENV API_SERVER_PORT=80
EXPOSE 80

ENTRYPOINT ["./supmap-gis"]