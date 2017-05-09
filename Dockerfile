FROM golang:1.7.3-alpine
ADD . /turbulence
WORKDIR /turbulence
ENV GOPATH /turbulence
RUN go build
ENTRYPOINT ["/bin/sh", "-c", "/turbulence/turbulence ${*}", "--"]
