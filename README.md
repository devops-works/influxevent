# Influxevent

[![Build Status](https://travis-ci.org/devops-works/influxevent.svg?branch=master)](https://travis-ci.org/devops-works/influxevent)
[![Go Report Card](https://goreportcard.com/badge/github.com/devops-works/influxevent)](https://goreportcard.com/report/github.com/devops-works/influxevent)
[![Maintainability](https://api.codeclimate.com/v1/badges/c4371a3fb883f200ccef/maintainability)](https://codeclimate.com/github/devops-works/influxevent/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/c4371a3fb883f200ccef/test_coverage)](https://codeclimate.com/github/devops-works/influxevent/test_coverage)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

**this is beta stuff**

Influxevent wraps commands and sends command result & timing to an influxDB server.

The influx entries contains the following values:

- *command duration* (tag `etype=event`)
- *command exit value* (tag `etype=event`)
- *command cpu usage* for per sample period (tag `etype=metric`)
- *command memory usage* for per sample period (tag `etype=metric`)

When invoking `influxevent`, you can pass:

- the measurement name (e.g. `cron`)
- a list of tag names and values (`tag1=val1,tag2=val2`); `host` tag is added
  automatically

Influxevent transparently proxies command's stdout/stderr and exit value.

## Usage

### Quick start

Using in a cron job:

```bash
0 1 * * * ./influxevent -timeout 300 -server https://influx.example.com:8086/ -db mydb -measurement cron -period 100 -tag command=backup -retry 3 -- /usr/local/bin/database_backup >> /var/log/backups.log 2>&1
```

then in influxdb:

```bash
> select * from cron
name: cron
time                           command cpu duration    etype  host  memory status
----                           ------- --- --------    -----  ----  ------ ------
2020-05-10T22:51:05.390134036Z backup      10.06035404 event  host1        0
2020-05-10T22:51:05.491496489Z backup  10              metric host1 733184 
2020-05-10T22:51:05.591504518Z backup  12              metric host1 733184 
2020-05-10T22:51:05.691497143Z backup  11              metric host1 733184 
2020-05-10T22:51:05.791500623Z backup  15              metric host1 733184 
...
>
```

Adding tags:

```bash
0 1 * * * ./bin/influxevent -server https://influx.example.com:8086/ -db mydb -measurement events --tags program=database_backup,db=foodb -- /usr/local/bin/database_backup foodb >> /var/log/backups.log 2>&1
```

### Arguments

General invocation: `influxevent [options] -- cmd args...`

Options:

- `-influx_url` (`$INFLUX_URL`): influxdb server URL (no events are send if not set)
- `-influx_db` (`$INFLUX_DB`): influxdb database (no events are send if not set)
- `-influx_user` (`$INFLUX_USER`): influxdb username (default: none)
- `-influx_pass` (`$INFLUX_PASS`): influxdb password (default: none)
- `-influx_measurement` (`$INFLUX_MEASUREMENT`): influxdb measurement (default: none, required when server is set)
- `-influx_tags` (`$INFLUX_TAGS`): comma-separated k=v pairs of influxdb tags (default: none, example: 'foo=bar,fizz=buzz')
- `-influx_retries` (`$INFLUX_RETRIES`): how many times we retry to send the event to influxdb(default: 3) (default 3)
- `-influx_timeout` (`$INFLUX_TIMEOUT`): timeout writing to influxdb in ms (default: 5000)
- `-influx_dryrun` (`$INFLUX_DRYRUN`): influxdb dry run (runs command and dumps influx datapoints instead of sending them to influxdb)
- `-timeout` (`$TIMEOUT`): command timeout (default: 0, no timeout)
- `-period` (`$PERIOD`): process consumption sample period (ms)
- `-verbose` (`$VERBOSE`): verbose execution
- `-version`: shows version

You can leverage environment variables in crontabs to have shorter cron
definitions. For instance:

```
INFLUX_URL=http://1.2.3.4:8086
INFLUX_DB=mydb
INFLUX_USER=user
INFLUX_PASS=pass
INFLUX_MEASUREMENT=cron
PERIOD=100

# Database Backup
0 1 * * * ./influxevent -tags "type=backup,engine=mysql,db=customers" -- /usr/local/bin/database_backup >> /var/log/backups.log 2>&1
0 6 * * 7 ./influxevent -tags "type=certbot,domain=example.org" -- /usr/bin/certbot certonly -n -d example.com -d www.example.com >> /var/log/certbot.log 2>&1
```

## Installing

### Binary

```bash
# YOLO
curl -sL https://github.com/devops-works/influxevent/releases/download/v0.5/influxevent-amd64-v0.5.gz -o - | gunzip > influxevent
chmod +x influxevent
sudo mv influxevent /usr/local/bin/influxevent
```

or if you have Go installed:

```bash
go install github.com/devops-works/influxevent
```

### Compiling

```bash
make
```

## Contributions

Welcomed !

## Licence

GPLv3