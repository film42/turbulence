FROM golang:1.11-alpine
ADD . /turbulence
WORKDIR /turbulence
ENV GOPATH /turbulence
RUN go build
ENTRYPOINT ["/turbulence/turbulence"]
