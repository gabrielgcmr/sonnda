# Etapa 1: build
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Copia go.mod, go.sum e baixa dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código e compila
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/api/main.go

# Etapa 2: runtime
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/api .
# Expõe a porta usada pelo Gin (por padrão 8080)
EXPOSE 8080
ENTRYPOINT ["./api"]
