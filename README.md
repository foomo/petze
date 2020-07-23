<img src="petze.png" width="300" height="300">

[![Travis CI](https://travis-ci.org/foomo/petze.svg?branch=master)](https://travis-ci.org/foomo/petze)

# Petze

**Petze monitors web sites and services**. It exports [prometheus.io](https://prometheus.io) metrics and writes a [logrus](https://github.com/Sirupsen/logrus) log.

One instance of petze is designed to monitor many services at little cost.

# Motivation

While there is a vast amount of monitoring solutions out there I still felt there was somthing simplistic missing, that would play nicely with [prometheus.io](https://prometheus.io), which I have come to appreciate very much.

# Status

As of June 2017 we have a first working version and we are trying to get the configs right - feedback and contributions are most welcome!

# Configuration

Petze is configured through a set of yaml configuration files. The path to folder containing the configuration is passed to petze as the first argument on the commandline.

```bash
$ petze path/to/petzconf
``` 

Take a look at a simple example config: https://github.com/foomo/petze/tree/master/exampleConfig

## Configuration layout

The configuration file for petze is called: **petze.yml**.
It provides information for the petze service, as well the configuration for your notifications.

## Main config file petze.yml

```yaml
# HTTP endpoint for prometheus metrics
address: server-name.net:8080

# optional: running with TLS
tls:
  address: server-name:8443
  cert: path/to/cert.pem
  key: path/to/key.pem

# optional: notification via slack webhooks
slack: https://hooks.slack.com/services/custom-parameters

# optional: configure SMTP notifications
smtp:
  server: smtp-relay.yourprovider.com
  user: you@mail.com
  pass: yourpassword
  port: 465
  from: replyto@mail.com
  to: usertonotify@mail.com 

# optional basic auth
basicauthfile: path/to/basic-auth-file
```

## Service configuration files

Any other files with a .yml suffix will be treated as service configurations. 
It is strongly encouraged to organize them in folder structures. 
These will be reflected in the service ids.

```yaml
endpoint: http://www.bestbytes.de
interval: 5m
tlswarning: 128h # overwrite the default warning of one week before expiry for this service
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
      - headers:
          "X-Test": "foo"
      - json-path:
        # this is a json path expression
        "$[0].product.images+":
        	min: 1
  - uri: "/another/path"
    check:
      - duration: 100ms
      - redirect: "https://myservice.com/asdf" # match the location for checking redirects
      - match-reply: "asdf" # match the raw response string

```

## SMTP Integration

You can now get notifications by Mail, all you need to provide is an SMTP server!
A summary email with all errors for a service will be generated, in case a check failed.

Add the following to your petze.yml:

```yaml
# configure SMTP notifications
smtp:
  server: smtp-relay.yourprovider.com
  user: you@mail.com
  pass: yourpassword
  port: 465
  from: replyto@mail.com
  # enter multiple emails to notify if desired
  to: 
    - usertonotify1@mail.com
    - usertonotify2@mail.com 
```

## Slack Integration

Using slack incoming webhooks, we can post messages to a slack channel
by simply creating a slack app and enabling the incoming webhook:
https://api.slack.com/messaging/webhooks

1) Add a new App to your Slack workspace
2) Go to the Apps Settings and enable Webhooks
3) Select a channel to notify and generate a new webhook URL

Then add your newly generated webhook URL to your petze.yml:

```yaml
slack: https://hooks.slack.com/services/custom-parameters
```

## Docker Usage

Prepare your config folder and move it to: /etc/petzconf.
The repository contains an example configuration in the _exampleConfig_ folder.

Then pull and start the container, mounting the config folder into the container:

```bash
$ docker pull foomo/petze
$ docker run -v /etc/petzconf:/etc/petzconf foomo/petze
```

Happy monitoring!
