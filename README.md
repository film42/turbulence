Turbulence
==========

[![Build Status](https://travis-ci.org/film42/turbulence.svg)](https://travis-ci.org/film42/turbulence)

An http/https proxy for navigating through rough air. This proxy is not full-featured. It doesn't support authentication, whitelisting or custom headers. It's small, transparent, and does as little as possible. Be sure to read through the code, it should take you less than 5 minutes.

### Installing

To build from source you can do:

```
$ go get github.com/film42/turbulence
$ ./bin/turbulence
```

Or you can grab the pre-built docker container:

```
$ docker run -p 26000:26000 -d film42/turbulence:latest --config test/config.json
```

### Configuring

Turbluence only allows the listen port, proxy username, and proxy password to be configured:

```
Usage of ./turbulence:
  -config string
        config file
  -password string
        password for proxy authentication
  -port int
        listen port (default 25000)
  -shutdown-timeout int
        seconds to wait while cleaning up for connections to finish (default 60)
  -strip-proxy-headers
        strip proxy headers from http requests (default true)
  -username string
        username for proxy authentication
```

Turbulence supports basic authentication. Other massive proxies like Squid support complex authentication, whitelisting and custom headers, but Turbulence doesn't try to solve these other problems. It's built to do as little as possible. If you _need_ authentication or the ability to modify headers, then you should use a different proxy, but if whitelisting is good enough, use `iptables` instead. Even if you use Squid's whitelisting, the proxy is still publicly accessible and will respond to proxy connections with an unauthorized error. If you're using a cloud provider like amazon or google, be sure to use their firewall tools.

If you want to use a sample config instead of passing your options as command line arguments, you may do so. Here's an
example config:

```
{
  "port": 9000,
  "strip_proxy_headers": true,
  "credentials": [
    {
      "username": "ron.swanson",
      "password": "g0ld5topsTheGovt"
    }
  ]
}
```

Save and run:

```
./turbulence --config config_for_ron.json
```

### Purpose

After using large proxies like Squid, I realized I was only using a very small subset of its features and when something would go wrong, I wasn't sure what why. After a few hours of trying to solve those problems, I looked up the proxy server RFC and realized how easy it would be to start from scratch and only add the most basic components required to proxy http and https requests. After first writing a functional proxy in ruby with nio4r, I realized go would be a much easier runtime to work with. Turbulence is the result of a few iterations of writing a toy proxy. Turbulence isn't nearly as battle tested as Squid, but it is very tiny and can fit into your head. I've hit Turbulence with pings and full proxy requests continuously for over two weeks (still running) and the memory utilization has been about 1/3 of Squid, but YMMV.

### Technical Details

Turbulence is a transparent http/https proxy, meaning it let's you proxy your http/https traffic without modifying any of the original request. This doesn't necessarily mean that Turbulence is anonymous. If your client sets an `X-Forwarded-For` header or any other request header that could hint the request is being proxied, it will be sent to the server. If you want your requests to look like they're not proxied, be sure that your client does not set any additional headers. While http requests can be modified by Turbulence without raising any red flags, it's not possible to modify https requests without man-in-the-middling your connection, which is detectable by any client that pins their certificate. For this reason, Turbulence does not manipulation http/https request and only proxies bytes between two tcp connections.

**HTTP Request**

When a client connects to the proxy using HTTP, the HTTP request will look like a regular HTTP request except for the first line which will contain the full host of the destination server:

```
GET https://server.example.com/articles HTTP/1.1
...
```

Turbulence connects to the host `server.example.com:80` and writes the http request without `https://server.example.com` prefixed to the request path. After this, Turbulence copies bytes between tcp connections until one of them hangs up. This allows Turbulence to support http keepalive for both http and https mode.


Try using curl to watch the request flow:

```
$ curl -v -x 0.0.0.0:25000 http://httpbin.org/headers http://httpbin.org/headers
```

**HTTPS Request**

When a client connects to the proxy using HTTPS, the client sends a `CONNECT` request to change the connection into socket mode for a TLS handshake.

```
CONNECT server.example.com:443 HTTP/1.1
...
```
Turbulence opens a tcp connection to `server.example.com:443` and responds to the client with a `200 Connection established` response to say everything is :ok_hand:. At this point, Turbulence copies bytes between the client and server until one of them hangs up. This allows Turbulence to supports http keepalive.

Try using curl to watch the request flow:

```
$ curl -v -x 0.0.0.0:25000 https://httpbin.org/headers https://httpbin.org/headers
```

### Example Client Request

To make an example proxied https connection with ruby's `Net::HTTP`, do the following:

```ruby
require "net/http"
uri = URI.parse("https://httpbin.org/headers")
proxy = Net::HTTP::Proxy("0.0.0.0", 25000)
client = proxy.start(uri.host, :use_ssl => true)
response = client.get(uri.request_uri)
puts response.body
```

Which will return the following:

```json
{
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip;q=1.0,deflate;q=0.6,identity;q=0.3",
    "Host": "httpbin.org",
    "User-Agent": "Ruby"
  }
}
```
