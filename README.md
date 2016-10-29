Turbulence
==========

An http/https proxy for navigating through rough air.

#### Usage

```
$ go run *.go
```

Example connection to google's homepage:
```
[info] 2016/10/29 14:13:00 turbulence.go:38: Prepare for takeoff...
[info] 2016/10/29 14:13:00 turbulence.go:47: Server started on :25000
[info] 2016/10/29 14:13:07 connection.go:20: [ed8b8d] Handling new connection.
[info] 2016/10/29 14:13:07 connection.go:34: [ed8b8d] Processing connection to: CONNECT www.google.com:443
[info] 2016/10/29 14:13:07 connection.go:20: [a3897a] Handling new connection.
[info] 2016/10/29 14:13:07 connection.go:34: [a3897a] Processing connection to: CONNECT ssl.gstatic.com:443
[info] 2016/10/29 14:13:07 connection.go:20: [74e997] Handling new connection.
[info] 2016/10/29 14:13:07 connection.go:34: [74e997] Processing connection to: CONNECT www.gstatic.com:443
[info] 2016/10/29 14:13:07 connection.go:20: [087b5f] Handling new connection.
[info] 2016/10/29 14:13:07 connection.go:34: [087b5f] Processing connection to: CONNECT apis.google.com:443
[info] 2016/10/29 14:13:14 connection.go:70: [087b5f] Connection closed.
[info] 2016/10/29 14:13:14 connection.go:70: [74e997] Connection closed.
[info] 2016/10/29 14:13:14 connection.go:70: [a3897a] Connection closed.
[info] 2016/10/29 14:13:14 connection.go:70: [ed8b8d] Connection closed.
```

#### Docker

```
docker run -p 25000:25000 -d film42/turbulence:latest
```
