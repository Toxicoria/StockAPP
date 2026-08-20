FROM golang:1.25-alpine AS builder
WORKDIR /app

# Capa de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el módulo de Go y compilar binario estático sin CGO
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
