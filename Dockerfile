# Start from a Go 1.25 base image
FROM golang:1.25-alpine AS build

WORKDIR /src

# Copy go.mod and go.sum early for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build your app (adjust as needed)
RUN go build -tags netgo -ldflags="-s -w" -o app cmd/server/main.go

# Use a minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=build /src/app .

# Expose port (default is 10000 on Render)
EXPOSE 10000

CMD ["./app"]
