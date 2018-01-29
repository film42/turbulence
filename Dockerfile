FROM golang:1.8.1-alpine
ADD . /turbulence
WORKDIR /turbulence
ENV GOPATH /turbulence
RUN go build
ENTRYPOINT ["/turbulence/turbulence"]
