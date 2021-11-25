FROM golang:alpine3.14 AS build_piuma
WORKDIR /app

RUN apk add --update --no-cache git build-base pkgconfig

ENV SOURCE_DIR=/go/src/github.com/piumaio/piuma
ADD . ${SOURCE_DIR}

RUN cd ${SOURCE_DIR} && \
  go mod download && \
  CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -X main.version=`git rev-parse HEAD`" -o app && \
  cp app /app

FROM alpine:3.14

# Install all required tools
RUN apk add --update --no-cache optipng jpegoptim libwebp libstdc++ dssim
RUN apk add libavif-apps --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community

WORKDIR /root/
COPY --from=build_piuma /app .
ENTRYPOINT ["./app"]
