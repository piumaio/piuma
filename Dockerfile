FROM golang:latest
WORKDIR /app
ENV SOURCE_DIR=/go/src/github.com/piumaio/piuma
ADD . ${SOURCE_DIR}

RUN cd ${SOURCE_DIR}; go get -u; CGO_ENABLED=0 GOOS=linux go build -o app; cp app /app

FROM alpine:3.10.2
RUN apk update

# Install OptiPNG
RUN apk add optipng

## Install JPEGOptim
RUN apk add jpegoptim

WORKDIR /root/
COPY --from=0 /app .
ENTRYPOINT ["./app"]
