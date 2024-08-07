# this is base image that builds all binaries.
# see app/cmd/<APP>/Dockerfile for individual app binary

# keep the golang version in sync with the .tool-version file
FROM golang:1.22.5 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN update-ca-certificates
RUN useradd -u 10001 scratchuser

WORKDIR /code

# vendor dir may not exist, thus the [r] trick.
# https://stackoverflow.com/questions/31528384/conditional-copy-add-in-dockerfile
COPY vendo[r] vendor

COPY go.mod go.sum ./
COPY docs docs
COPY db db
COPY app app

RUN go build -o=/go/bin ./app/cmd/...
