FROM golang:alpine3.15 AS build_piuma
WORKDIR /app

RUN apk add --update --no-cache git build-base pkgconfig

ENV SOURCE_DIR=/go/src/github.com/piumaio/piuma
ADD . ${SOURCE_DIR}

RUN cd ${SOURCE_DIR} && \
  go mod download && \
  CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -X main.version=`git rev-parse HEAD`" -o app && \
  cp app /app

FROM alpine:3.15

ENV PORT=8080
ENV MEDIA_DIR=~/.piuma/media
ENV TIMEOUT=0
ENV HTTP_CACHE_TTL=3600
ENV HTTP_CACHE_PURGE_INTERVAL=3600
ENV WORKERS=4
ENV DOMAINS=""


# Install all required tools
RUN apk add --update --no-cache optipng jpegoptim libwebp libstdc++ dssim libavif-apps

WORKDIR /root/
COPY --from=build_piuma /app .
ENTRYPOINT ./app -port ${PORT} -mediapath ${MEDIA_DIR} -timeout ${TIMEOUT} -httpCacheTTL ${HTTP_CACHE_TTL} -httpCachePurgeInterval ${HTTP_CACHE_PURGE_INTERVAL} -workers ${WORKERS} -domains "${DOMAINS}"
