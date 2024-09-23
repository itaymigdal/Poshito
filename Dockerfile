FROM golang:1.23.1-alpine

COPY . /Poshito

RUN apk add --no-cache git upx python3

WORKDIR /Poshito/Poshito/Agent

RUN go mod download && go install mvdan.cc/garble@latest

