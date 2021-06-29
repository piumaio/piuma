#
# Here libheif is built from scratch to use rav1e and dav1d respectively for 
# encoding and decoding as they seems to be faster than libaom which is used 
# in default package.
#

FROM alpine:3.14 AS heif_build
RUN apk add --update --no-cache git build-base libjpeg-turbo-dev libpng-dev dav1d-dev cmake
RUN apk add rav1e-dev --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing

WORKDIR /heif_build/
RUN git clone --branch v1.11.0 https://github.com/strukturag/libheif.git
ENV PKG_CONFIG_PATH="/usr/lib/pkgconfig/dav1d.pc:/usr/lib/pkgconfig/rav1e.pc:${PKG_CONFIG_PATH}"
ENV CPPFLAGS="-O2"
ENV CXXFLAGS="-O2"
RUN cd libheif && cmake . -DWITH_LIBDE265=OFF -DWITH_X265=OFF -DWITH_EXAMPLES=ON -DBUILD_SHARED_LIBS=OFF && make

FROM golang:alpine3.14 AS build_piuma
WORKDIR /app

RUN apk add --update --no-cache git build-base pkgconfig

ENV SOURCE_DIR=/go/src/github.com/piumaio/piuma
ADD . ${SOURCE_DIR}

RUN cd ${SOURCE_DIR} && \
  go mod download && \
  CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o app && \
  cp app /app

FROM alpine:3.14

# Install all required tools
RUN apk add --update --no-cache optipng jpegoptim libwebp dav1d-dev libstdc++ dssim
RUN apk add rav1e-dev --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing
COPY --from=heif_build /heif_build/libheif/examples/heif-* /usr/bin/

WORKDIR /root/
COPY --from=build_piuma /app .
ENTRYPOINT ["./app"]
