FROM golang:1.7.3-alpine
ADD . /turbulence
WORKDIR /turbulence
ENV GOPATH /turbulence
RUN go build
CMD ["/turbulence/turbulence"]
