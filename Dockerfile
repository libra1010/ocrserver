FROM golang:1.12
ENV GO111MODULE=on

LABEL maintainer="otiai10 <otiai10@gmail.com>"

RUN apt-get -qq update \
  && apt-get install -y \
    libleptonica-dev \
    libtesseract-dev \
    tesseract-ocr

# Load languages
RUN apt-get install -y -qq \
  tesseract-ocr-jpn    \
  tesseract-ocr-eng    \
  tesseract-ocr-chi-sim


RUN go build

ENV PORT=8080
CMD $GOPATH/bin/ocrserver
