FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY tern.go conn.go webtransport.go agents-guide.md ./
COPY crypto/ crypto/
COPY cmd/tern/ cmd/tern/
RUN CGO_ENABLED=0 go build -o /tern ./cmd/tern

FROM alpine:3.21
COPY --from=build /tern /tern
EXPOSE 443/udp
CMD ["/tern"]
