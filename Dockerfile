# Etapa 1: build
FROM golang:1.24-alpine AS build
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
# Copia go.mod, go.sum e baixa dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código e compila
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/sonnda-api ./cmd/api

# Etapa 2: runtime
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /bin/sonnda-api /app/sonnda-api

# Expõe a porta usada pelo Gin (por padrão 8080)
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/sonnda-api"]
