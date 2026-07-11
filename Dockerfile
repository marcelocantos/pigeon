FROM golang:1.25-alpine AS build
WORKDIR /src

# Build deps for libsodium's autotools chain + cgo.
RUN apk add --no-cache build-base autoconf automake libtool bash

# Build the vendored static libsodium that cwire links against.
# Cached when the submodule SHA and build.sh are unchanged.
COPY c/vendor/build.sh c/vendor/build.sh
COPY c/vendor/github.com/jedisct1/libsodium c/vendor/github.com/jedisct1/libsodium
RUN bash c/vendor/build.sh libsodium

COPY go.mod go.sum ./
RUN go mod download

COPY *.go agents-guide.md ./
COPY crypto/ crypto/
COPY protocol/ protocol/
COPY qr/ qr/
COPY cwire/ cwire/
COPY dist/ dist/
COPY cmd/pigeon/ cmd/pigeon/

# VERSION / COMMIT are injected by CI (`flyctl deploy --build-arg`) so
# GET /status and --version report what is actually running. Local
# docker builds default to "dev" with an empty commit.
ARG VERSION=dev
ARG COMMIT=
RUN CGO_ENABLED=1 go build \
    -ldflags "-X main.version=${VERSION} -X github.com/marcelocantos/pigeon.Commit=${COMMIT}" \
    -o /pigeon ./cmd/pigeon

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && mkdir -p /data/certmagic
COPY --from=build /pigeon /pigeon
EXPOSE 443/udp 443/tcp 4433/udp
CMD ["/pigeon"]
