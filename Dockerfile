FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go agents-guide.md ./
COPY crypto/ crypto/
COPY protocol/ protocol/
COPY qr/ qr/
COPY cmd/tern/ cmd/tern/
RUN CGO_ENABLED=0 go build -o /tern ./cmd/tern

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && mkdir -p /data/certmagic
COPY --from=build /tern /tern
EXPOSE 443/udp 443/tcp 4433/udp
CMD ["/tern"]
