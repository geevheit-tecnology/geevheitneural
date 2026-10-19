# Stage 1: Build da aplicação Go
FROM golang:alpine AS builder

WORKDIR /app

# Instala certificados CA
RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/geevheit-api \
    ./cmd/geevheit-api

# Stage 2: Imagem de execução mínima e segura (distroless)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/geevheit-api /geevheit-api

USER nonroot:nonroot

EXPOSE 8085

ENTRYPOINT ["/geevheit-api"]
