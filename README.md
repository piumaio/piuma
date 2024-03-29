<p align="center"><img src="https://raw.githubusercontent.com/astagi/mystatics/master/piuma/Piuma_rounded_1.png" width='192' height="183" /></p>


# Piuma    
[![](https://images.microbadger.com/badges/version/piumaio/piuma.svg)](https://microbadger.com/images/piumaio/piuma "Get your own version badge on microbadger.com")
[![Build Status](https://travis-ci.org/piumaio/piuma.svg?branch=master)](https://travis-ci.org/piumaio/piuma) [![Coverage Status](https://img.shields.io/codecov/c/github/piumaio/piuma.svg)](https://codecov.io/gh/piumaio/piuma)


Simple and fast image optimizer service you can host on your machine

## Install

```
go get github.com/piumaio/piuma
```

## Requirements

Since this project automates two applications, you will need them to be installed on your machine for us to be able to reach them:

- [OptiPNG](http://optipng.sourceforge.net/)
- [jpegoptim](https://github.com/tjko/jpegoptim)
- [libwebp](https://developers.google.com/speed/webp/download)
- [libheif](https://github.com/strukturag/libheif) along with its tools `heif-enc` and `heif-convert`

Also, [dssim](https://github.com/kornelski/dssim), used for adaptive image quality, is suggested but not required.

## Run

```
piuma
```

You can also change the default `port` (`8080` by default) and `mediapath`, type

```
piuma --help
```

for more info.

# Running with Docker

Use the following command to build the Docker image from the root folder:

```
docker build -t piuma .
```

Next, you can run the image and provide the port and the mediapath where the optimized images will be stored:
```
docker run -p 8080:8080 -v $PWD:/data piuma -mediapath /data
```

Above command will run Piuma on ```http://localhost:8080``` and it's going to store all optimized images in the current directory (```$PWD```).

## Usage

```
https://yourpiumahost/Options/Image_URL
```

Where options are values separated by `_`

```
width_height_quality
```

or

```
width_height_quality:optional_image_format
```

Where `quality` is a value between 0 and 100.

To get your image resized to 100 x 100:

```
https://yourpiumahost/100_100/<Image_URL>
```

If you want to specify only the `width`, you'll get a new image keeping the ratio:

```
https://yourpiumahost/100/<Image_URL>
```

If you want to specify only the `height`

```
https://yourpiumahost/0_100/<Image_URL>
```

If you want to convert the image to a specific format add a `:image_extension` 
where `image_extension` can be one of the following:

* `jpg` or `jpeg` for [JPEG](https://en.wikipedia.org/wiki/JPEG)
* `png` for [PNG](https://en.wikipedia.org/wiki/Portable_Network_Graphics)
* `webp` for [WebP](https://en.wikipedia.org/wiki/WebP)
* `webp_lossless` same as `webp` but with lossless conversion
* `avif` for [AVIF](https://en.wikipedia.org/wiki/AV1#AV1_Image_File_Format_(AVIF))
* `auto` that chooses the best supported image format by parsing the `Accept` request header
    * If you add a colon followed by a comma-separated list of extension you can pass a list of allowed extension
     (e.g. `auto:webp,jpg,png` will only select webp, jpeg and png extensions)

```
https://yourpiumahost/0_0_100:webp/<Image_URL>
```

#### Adaptive quality
Also you can add an `a` after quality value (e.g. `0_0_75a`), that means that quality will be choosed following a DSSIM value generated by this expression: `100-quality_value/10000`.

Basically it can be used whenever you want to have the same perceived quality among multiple image formats, but requires more time because image needs to be converted more times for searching an optimal quality value.

## Running tests
To run the unit tests, change to the directory with tests (files ending with ```_test.go``` contain unit tests) and run:

```
go test -v ./...
```
