FROM golang:1.9.1-stretch
LABEL name piuma
RUN apt update
RUN apt install -y pngquant jpegoptim
WORKDIR /go/src/app
COPY . .
RUN go-wrapper download
RUN go-wrapper install
RUN go build -v
EXPOSE 8080
CMD ["app"]
