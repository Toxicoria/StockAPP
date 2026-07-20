FROM golang:1.25-alpine AS builder
WORKDIR /app

# Primero solo las dependencias: esta capa se cachea mientras no cambie go.mod/go.sum
COPY go.mod go.sum ./
RUN go mod download

# Después el código (solo .go: el CSV, los .sql y los binarios no van a la imagen)
COPY *.go ./
COPY db/ db/
COPY pkg/ pkg/
RUN go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]
