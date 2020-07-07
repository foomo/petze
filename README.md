[![Travis CI](https://travis-ci.org/foomo/petze.svg?branch=master)](https://travis-ci.org/foomo/petze)

- add support to match raw response against string or regex (eg to check robots file)
- add example config directory 
- add support to check health of raw TCP / UDP services

# Petze

**Petze monitors web sites and services**. It exports [prometheus.io](https://promtheus.io) metrics and writes a [logrus](https://github.com/Sirupsen/logrus) log.

One instance of petze is designed to monitor many services at little cost.

# Motivation

While there is a vast amount of monitoring solutions out there I still felt there was somthing simplistic missing, that would play nicely with [prometheus.io](https://promtheus.io), which I have come to appreciate very much.

# Status

As of June 2017 we have a first working version and we are trying to get the configs right - feedback is most welcome.

# Configuration

Petze is configured through a set of yaml configuration files. The config folder name can be passed to petze, if not it will look for the configuration in 

## Configuration layout

Petze is configured from files in a configuration folder.

## Main config file petze.yml

```yaml
# 
# optional http address if you want to run without tls
address: server-name.net:8080
# running on tls
tls:
  address: server-name:8443
  cert: path/to/cert.pem
  key: path/to/key.pem
# optional basic auth
basicauthfile: path/to/basic-auth-file

```

## Service configurations

Any other files with a .yml suffix will be treated as service configurations. It is strongly encouraged to organize them in folder structures. These will be refected in the service ids.





```yaml
---
endpoint: http://www.bestbytes.de
interval: 5m
# run requests in a session, with cookies
session:
  - uri: "/"
    comment: home page visit
    check:
      - statuscode: 200
      - duration: 200ms
      - goquery:
      	  ".body div.test":
      	  min: 3
  
  - method: POST
    comment: this is how you perform XHR requests  
    uri: "/path/to/a/rest/service?foo=bar"
    content-type: application/json
    headers:
      "X-Test": ["foo"]
    data:
      foo: bar
    check:
      - content-type: application/json
      - duration: 100ms
      - header:
          "X-Test": "foo"
      - json-path:
        # this is a json path expression
        "$[0].product.images+":
        	min: 1
  - uri: "/another/path"
    check:
      - duration: 100ms

```

