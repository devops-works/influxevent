# Influxevent

[![Build Status](https://travis-ci.org/devops-works/influxevent.svg?branch=master)](https://travis-ci.org/devops-works/influxevent)
[![Go Report Card](https://goreportcard.com/badge/github.com/devops-works/influxevent)](https://goreportcard.com/report/github.com/devops-works/influxevent)
[![Maintainability](https://api.codeclimate.com/v1/badges/c4371a3fb883f200ccef/maintainability)](https://codeclimate.com/github/devops-works/influxevent/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/c4371a3fb883f200ccef/test_coverage)](https://codeclimate.com/github/devops-works/influxevent/test_coverage)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

**this is beta stuff**

Influxevent wraps commands and sends command result & timing to an influxDB server.

The event contains the following values:

- *command duration*
- *command exit value*

When invoking `influxevent`, you can pass:

- the measurement name (e.g. `events`)
- a list of tag names and values

## Usage

### Quick start

Using in a cron job:

```bash
0 1 * * * ./influxevent -timeout 300 -server https://influx.example.com:8086/ -db mydb -measurement events -retry 3 /usr/local/bin/database_backup >> /var/log/backups.log 2>&1
```

then in influxdb:

```bash
> select last(*) from events;
name: events
time last_duration last_status
---- ------------- -----------
0    5.001956      0
> 
```

Adding tags:

```bash
0 1 * * * ./bin/influxevent -server https://influx.example.com:8086/ -db mydb -measurement events --tags program=database_backup,db=foodb /usr/local/bin/database_backup foodb >> /var/log/backups.log 2>&1
```

### Arguments

General invocation: `influxevent [options] cmd args...`

Options:

- `-db`: influxdb database (no events are send if not set)
- `-measurement`: influxdb measurement (default: none, required when server is set)
- `-pass`: influxdb password (default: none)
- `-retry`: how many times we retry to send the event to influxdb(default: 3) (default 3)
- `-server`: influxdb server URL (no events are send if not set)
- `-tags`: comma-separated k=v pairs of influxdb tags (default: none, example: 'foo=bar,fizz=buzz')
- `-timeout`: command timeout (default: 0, no timeout)
- `-influxtimeout`: timeout writing to influxdb in ms (default: 2000)
- `-user`: influxdb username (default: none)
- `-version`: shows version

## Installing

### Binary

```bash
# YOLO
curl -sL https://github.com/devops-works/influxevent/releases/download/v0.1/influxevent-amd64-v0.1.gz -o - | gunzip > influxevent
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