FROM golang:1.8.1-alpine
ADD . /turbulence
WORKDIR /turbulence
ENV GOPATH /turbulence
RUN go build
ENTRYPOINT ["/bin/sh", "-c", "/turbulence/turbulence ${*}", "--"]
