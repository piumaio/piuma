FROM golang:latest
WORKDIR /app
ENV SOURCE_DIR=/go/src/github.com/lotrekagency/piuma
ADD . ${SOURCE_DIR}

RUN cd ${SOURCE_DIR}; go get -u; CGO_ENABLED=1 GOOS=linux go build -o app; cp app /app

FROM ubuntu
RUN apt update
RUN apt install -y ca-certificates pngquant jpegoptim
WORKDIR /root/
COPY --from=0 /app .
ENTRYPOINT ["./app"]