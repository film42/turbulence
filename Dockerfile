FROM golang:1.11-alpine as builder
ADD . /turbulence
WORKDIR /turbulence
ENV GOPATH /turbulence
RUN CGO_ENABLED=0 go build

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /turbulence/turbulence /
ENTRYPOINT ["/turbulence"]
