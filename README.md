# Influxevent

**this is alpha stuff - there aren't even tests :/**

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

then in fluxdb:

```bash
> select last(*) from events;
name: events
time last_duration last_status
---- ------------- -----------
0    5.001956      0
> 
```

### Arguments

General invocation: `influxevent [options] cmd args...`

Options:

- -db: influxdb database (no events are send if not set)
- -measurement: influxdb measurement (default: none, required when server is set)
- -pass: influxdb password (default: none)
- -retry: how many times we retry to send the event to influxdb(default: 3) (default 3)
- -server: influxdb server URL (no events are send if not set)
- -tags: comma-separated k=v pairs of influxdb tags (default: none, example: 'foo=bar,fizz=buzz')
- -timeout: command timeout (default: 0, no timeout)
- -user: influxdb username (default: none)
- -version: shows version

## Installing

### Binary

```bash
curl - ...
```

or if you have Go installed:

```bash
go install github.com/devops-works/influxevent
```

### Compiling

```bash
make
```

## Licence

DWTFYWT Public Licence