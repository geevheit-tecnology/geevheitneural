# Stage 1: Build da aplicação Go
FROM golang:alpine AS builder

WORKDIR /app

# Git é chamado pelo LocalPublisher para versionar os arquivos gerados.
RUN apk add --no-cache ca-certificates tzdata git

# Empacota apenas o executável, templates e bibliotecas exigidas pelo Git.
# Inclui o loader musl para funcionar na imagem distroless sem shell.
RUN mkdir -p /git-runtime/usr/bin /git-runtime/usr/share \
    && cp /usr/bin/git /git-runtime/usr/bin/git \
    && cp -a /usr/share/git-core /git-runtime/usr/share/ \
    && ldd /usr/bin/git | awk '{ for (i=1; i<=NF; i++) if ($i ~ /^\//) print $i }' | \
       while read -r lib; do \
         mkdir -p "/git-runtime$(dirname "$lib")"; \
         cp -L "$lib" "/git-runtime$lib" || exit 1; \
       done \
    && mkdir -p /git-runtime/dev \
    && touch /git-runtime/dev/null \
    && chroot /git-runtime /usr/bin/git --version \
    && rm /git-runtime/dev/null \
    && rmdir /git-runtime/dev

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
COPY --from=builder /git-runtime/ /

USER nonroot:nonroot

EXPOSE 8085

ENTRYPOINT ["/geevheit-api"]
