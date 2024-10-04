FROM golang:1.23.1-alpine

COPY . /Poshito

RUN apk add --no-cache git upx python3 mingw-w64-gcc \
&& go env -w CC=x86_64-w64-mingw32-gcc

WORKDIR /Poshito/Poshito/Agent

RUN go mod download \
&& go install mvdan.cc/garble@latest
