
# Piuma    [![Build Status](https://travis-ci.org/lotrekagency/piuma.svg?branch=master)](https://travis-ci.org/lotrekagency/piuma) [![Coverage Status](https://img.shields.io/codecov/c/github/lotrekagency/piuma.svg)](https://codecov.io/gh/lotrekagency/piuma)

### Simple and fast image optimizer server you can host on your machine
<img src="https://raw.githubusercontent.com/astagi/mystatics/master/piuma/Piuma_rounded_1.png" width='192' height="183" />

## Install

    $ go get github.com/lotrekagency/piuma

## Requirements

Since this project automates two applications, you will need them to be installed on your machine for us to be able to reach them:

- [pngquant](https://pngquant.org/)
- [jpegoptim](https://github.com/tjko/jpegoptim)

## Run

    $ piuma

You can also change the default `port` and `mediapath`, type

    $ piuma --help

for more info.

## Usage

    https://yourpiumahost/Options/Image_URL

Where options are values separated by `_`

    width_height_quality

Where `quality` is a value between 0 and 100.

To get your image resized to 100 x 100:

    https://yourpiumahost/100_100/Image URL

If you want to specify only the `width`, you'll get a new image keeping the ratio:

    https://yourpiumahost/100/Image URL

If you want to specify only the `height`

    https://yourpiumahost/0_100/Image URL

## Running tests
To run the unit tests, change to the directory with tests (files ending with ```_test.go``` contain unit tests) and run: 
 
    $ go test -v ./...
 
