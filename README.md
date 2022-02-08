# PlanB: a distributed HTTP and websocket proxy

[![Linter](https://github.com/vinayb21/planb/workflows/Linter/badge.svg)](https://github.com/vinayb21/planb/actions?query=workflow%3ALinter)
[![Tests](https://github.com/vinayb21/planb/workflows/Tests/badge.svg)](https://github.com/vinayb21/planb/actions?query=workflow%3ATests)
[![Coverage Status](https://coveralls.io/repos/github/jandelgado/golang-ci-template-github-actions/badge.svg?branch=master)](https://coveralls.io/github/vinayb21/planb?branch=master)

## What Is It?

PlanB is a HTTP and websocket proxy backed by Redis and inspired by
[Hipache](https://github.com/hipache/hipache).

It aims to be fully compatible with Hipache when Redis is used as a backend.
The same format is used for all keys stored in Redis so migrating from Hipache
to PlanB should be completely seamless. The process should be as simple as
replacing Hipache's executable for PlanB.

## Scripts

- `make start`: Build and run the planb binary
- `make build-docker` : Build the latest docker image with image name: planb and tag: latest
- `make clean` : Remove planb binary
- `make release` : Create the docker image and push to vm-provider project with the latest tag of the git release for prod
- `make release-dev` : Create the docker image and push to vm-provider-dev project with the latest tag of the git release for dev
- `make test` : Runs the tests and updates the coverage report

## Start-up flags

The following flags are available for configuring PlanB on start-up:

- `--listen value, -l value`: Address to listen (default: "0.0.0.0:8989")
- `--tls-listen value`: Address to listen with tls
- `--tls-preset value`: Preset containing supported TLS versions and cyphers, according
  to https://wiki.mozilla.org/Security/Server_Side_TLS. Possible
  values are [modern, intermediate, old] (default: "modern")
- `--metrics-address value`: Address to expose prometheus /metrics
- `--load-certificates-from value`: Path where certificate will found. If value equals 'redis'
  certificate will be loaded from redis service. (default:
  "redis")
- `--read-redis-network value`: Redis address network, possible values are "tcp" for tcp
  connection and "unix" for connecting using unix sockets
  (default: "tcp")
- `--read-redis-host value`: Redis host address for tcp connections or socket path for unix
  sockets (default: "127.0.0.1")
- `--read-redis-port value`: Redis port (default: 6379)
- `--read-redis-sentinel-addrs value`: Comma separated list of redis sentinel addresses
- `--read-redis-sentinel-name value`: Redis sentinel name
- `--read-redis-password value`: Redis password
- `--read-redis-db value`: Redis database number (default: 0)
- `--write-redis-network value`: Redis address network, possible values are "tcp" for tcp
  connection and "unix" for connecting using unix sockets
  (default: "tcp")
- `--write-redis-host value`: Redis host address for tcp connections or socket path for unix
  sockets (default: "127.0.0.1")
- `--write-redis-port value`: Redis port (default: 6379)
- `--write-redis-sentinel-addrs value`: Comma separated list of redis sentinel addresses
- `--write-redis-sentinel-name value`: Redis sentinel name
- `--write-redis-password value`: Redis password
- `--write-redis-db value`: Redis database number (default: 0)
- `--access-log value`: File path where access log will be written. If value equals
  'syslog' log will be sent to local syslog. The value 'none' can
  be used to disable access logs. (default: "./access.log")
- `--request-timeout value`: Total backend request timeout in seconds (default: 30)
- `--dial-timeout value`: Dial backend request timeout in seconds (default: 10)
- `--client-read-timeout value`: Maximum duration for reading the entire request, including the
  body (default: 0s)
- `--client-read-header-timeout value`: Amount of time allowed to read request headers (default: 0s)
- `--client-write-timeout value`: Maximum duration before timing out writes of the response
  (default: 0s)
- `--client-idle-timeout value`: Maximum amount of time to wait for the next request when
  keep-alives are enabled (default: 0s)
- `--dead-backend-time value`: Time in seconds a backend will remain disabled after a network
  failure (default: 30)
- `--flush-interval value`: Time in milliseconds to flush the proxied request (default: 10)
- `--request-id-header value`: Header to enable message tracking
- `--active-healthcheck`: Enable active healthcheck on dead backends once they are marked
  as dead. Enabling this flag will result in dead backends only
  being enabled again once the active healthcheck routine is able
  to reach them.
- `--engine value`: Reverse proxy engine, options is 'native'
- `--backend-cache`: Enable caching backend results for 2 seconds. This may cause
  temporary inconsistencies.
- `--help, -h`: show help
- `--version, -v`: print the version

The `--read-redis-*` flags refer to the Redis server used for read-only
operations (reading the backends for each frontend).

The `--write-redis-*` flags refer to the Redis server used for write operations
(marking and publishing dead backends).

Separating the read and write servers is not mandatory but is useful for
improving performance. A common scenario is having a slave Redis server on
localhost configured as `--read-redis` and a remote Redis master configured as
`--write-redis`.

## Features

* Load-Balancing
* Dead Backend Detection
* Dynamic Configuration
* WebSocket
* TLS

## VHOST Configuration

The configuration is managed by **Redis** that makes possible
to update the configuration dynamically and gracefully while
the server is running, and have that state shared across workers
and even across instances.

Let's take an example to proxify requests to 2 backends for the hostname
`www.tsuru.io`. The 2 backends IP are `192.168.0.42` and `192.168.0.43` and
they serve the HTTP traffic on the port `80`.

`redis-cli` is the standard client tool to talk to Redis from the terminal.

Follow these steps:

### Create the frontend:

```
$ redis-cli rpush frontend:www.tsuru.io mywebsite
(integer) 1
```

The frontend identifer is `mywebsite`, it could be anything.

### Add the 2 backends:

```
$ redis-cli rpush frontend:www.tsuru.io http://192.168.0.42:80
(integer) 2
$ redis-cli rpush frontend:www.tsuru.io http://192.168.0.43:80
(integer) 3
```

### Review the configuration:

```
$ redis-cli lrange frontend:www.tsuru.io 0 -1
1) "mywebsite"
2) "http://192.168.0.42:80"
3) "http://192.168.0.43:80"
```

### TLS Configuration using redis (optional)

```
$ redis-cli -x hmset tls:www.tsuru.io certificate < server.crt
$ redis-cli -x hmset tls:www.tsuru.io key < server.key

$ redis-cli -x hmset tls:*.tsuru.com certificate < wildcard.crt
$ redis-cli -x hmset tls:*.tsuru.io key < wildcard.key
```

### TLS Configuration using FS (optional)

create directory following this structure
```
cd certficates
ls
*.domain-wildcard.com.key
*.domain-wildcard.com.crt
absolute-domain.key
absolute-domain.crt
```

While the server is running, any of these steps can be
re-run without messing up with the traffic.

## Debbugging and Troubleshooting

One way to debug/toubleshoot planb is by analyzing the running goroutines.

Planb is able to handle the USR1 signal to dump goroutines in its execution
screen:

```
$ kill -s USR1 <planb-PID>
```
