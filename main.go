package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type InfluxServer struct {
	db      string
	url     string
	user    string
	pass    string
	retries int
}

type InfluxPoint struct {
	measurement string
	tags        string
	duration    float64
	status      int
}

func main() {
	var timeout = flag.Float64("timeout", 0, "command timeout (default: 0, no timeout")
	var influxServer = flag.String("server", "", "influxdb server URL (no events are send if not set)")
	var influxDB = flag.String("db", "", "influxdb database (no events are send if not set)")
	var influxUser = flag.String("user", "", "influxdb username (default: none)")
	var influxPass = flag.String("pass", "", "influxdb password (default: none)")
	var influxMeasurement = flag.String("measurement", "", "influxdb measurement (default: none, required when server is set)")
	var influxTags = flag.String("tags", "", "comma-separated k=v pairs of influxdb tags (default: none, example: 'foo=bar,fizz=buzz')")
	var influxRetry = flag.Int("retry", 3, "how many times we retry to send the event to influxdb(default: 3)")

	flag.Parse()

	start := time.Now()
	err := executeCommand(flag.Args(), *timeout)
	duration := time.Since(start).Seconds()

	exitStatus := 0

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			}
		} else {
			log.Fatalf("error running command: %v", err)
		}
	}

	if exitStatus == -1 {
		log.Printf("command has been killed due to timeout")
	}

	if *influxServer == "" || *influxDB == "" {
		log.Printf("not writing point to influx since database or url is not set (%d)", exitStatus)
		os.Exit(exitStatus)
	}

	inf := InfluxServer{
		db:      *influxDB,
		url:     *influxServer,
		user:    *influxUser,
		pass:    *influxPass,
		retries: *influxRetry,
	}

	pt := InfluxPoint{
		measurement: *influxMeasurement,
		tags:        *influxTags,
		duration:    duration,
		status:      exitStatus,
	}

	err = logInfluxDB(inf, pt)
	if err != nil {
		log.Printf("unable to write to influxdb: %v", err)
		os.Exit(1)
	}
}

func (p InfluxPoint) String() string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	influxString := fmt.Sprintf("%s,host=%s", p.measurement, host)
	influxString = strings.Join([]string{influxString, p.tags}, ",")
	influxString = fmt.Sprintf("%s duration=%f,status=%d", influxString, p.duration, p.status)
	return influxString
}

func logInfluxDB(server InfluxServer, point InfluxPoint) error {
	buf := bytes.NewBufferString(point.String())
	// fmt.Printf("influxString: %s\n", influxString)

	uri := fmt.Sprintf("%s/write?db=%s", server.url, server.db)

	// Dangerous; shoud use url encoding
	if server.user != "" {
		uri += fmt.Sprintf("&u=%s&p=%s", server.user, server.pass)
	}

	resp, err := http.Post(uri, "application/x-www-form-urlencoded", buf)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("unable to write to influxdb server %s, got reponse: %s", server.url, resp.Status)
	}
	return nil
}

func executeCommand(args []string, timeout float64) error {
	var cmd *exec.Cmd
	var ctx context.Context
	var cancel context.CancelFunc

	if timeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	cmd.Env = append(os.Environ())

	sout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	serr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	scanOut := bufio.NewScanner(sout)

	go func() {
		for scanOut.Scan() {
			fmt.Printf("out: %s\n", scanOut.Text())
		}
	}()

	scanErr := bufio.NewScanner(serr)

	go func() {
		for scanErr.Scan() {
			fmt.Printf("err: %s\n", scanErr.Text())
		}
	}()

	return cmd.Run()
}
