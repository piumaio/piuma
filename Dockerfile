FROM golang:alpine
WORKDIR /app
ENV SOURCE_DIR=/go/src/github.com/piumaio/piuma
ADD . ${SOURCE_DIR}

RUN apk add --update --no-cache git

RUN cd ${SOURCE_DIR} && \
  go get -u && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app && \
  cp app /app

FROM alpine:3.10.2

# Install OptiPNG and JPEGOptim
RUN apk add --update --no-cache optipng jpegoptim

WORKDIR /root/
COPY --from=0 /app .
ENTRYPOINT ["./app"]
