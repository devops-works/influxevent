package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/crgimenes/goconfig"
	_ "github.com/crgimenes/goconfig/yaml"

	"github.com/struCoder/pidusage"
)

type config struct {
	Timeout float64 `yaml:"timeout" cfg:"timeout" cfgDefault:"0" cfgHelper:"command timeout (s); 0 to disable"`
	Period  float64 `yaml:"period" cfg:"period" cfgHelper:"process consumption sample period (ms)"`
	Verbose bool    `yaml:"verbose" cfg:"verbose" cfgHelper:"verbose execution"`
	Influx  influx  `yaml:"influxdb" cfg:"influx" cfgHelper:"influxdb parameters"`
	command []string
}

type influx struct {
	DB          string  `yaml:"db" cfg:"db" cfgHelper:"influxdb database"`
	URL         string  `yaml:"url" cfg:"url" cfgHelper:"influxdb URL"`
	User        string  `yaml:"user" cfg:"user" cfgHelper:"influxdb user"`
	Pass        string  `yaml:"pass" cfg:"pass" cfgHelper:"influxdb password"`
	Measurement string  `yaml:"measurement" cfg:"measurement" cfgHelper:"influxdb measurement name"`
	Tags        string  `yaml:"tags" cfg:"tags" cfgHelper:"comma-separated influxdb tags (e.g. foo=bar,fizz=buzz)"`
	Retries     int     `yaml:"retries" cfg:"retries" cfgDefault:"3" cfgHelper:"influxdb retries when writing"`
	Timeout     float64 `yaml:"timeout" cfg:"timeout" cfgDefault:"5000" cfgHelper:"influxdb writing timeout (ms)"`
	DryRun      bool    `yaml:"dryrun" cfg:"dryrun" cfgDefault:"false" cfgHelper:"influxdb dry run (runs command but does not log to influxdb)"`
}

type point struct {
	measurement string
	tags        string
	values      map[string]float64
	timestamp   time.Time
}

// Version from git sha1/tags
var Version string

func anyInSlice(needles []string, stack []string) bool {
	for _, needle := range needles {
		if slicePosition(needle, stack) != -1 {
			return true
		}
	}

	return false
}

func slicePosition(needle string, stack []string) int {
	for i, item := range stack {
		if item == needle {
			return i
		}
	}

	return -1
}

func main() {
	if anyInSlice([]string{"-version", "--version"}, os.Args) {
		fmt.Printf("%s %s\n", os.Args[0], Version)
		os.Exit(0)
	}
	cfg := config{}

	// step 3: Pass the instance pointer to the parser
	err := goconfig.Parse(&cfg)
	if err != nil {
		log.Printf("unable to parse arguments: %+v", err)
		return
	}

	if cfg.Influx.Retries <= 0 {
		cfg.Influx.Retries = 3
	}

	pos := slicePosition("--", os.Args)
	if pos == -1 || pos > len(os.Args)-2 {
		log.Printf("error: no command specified, use: influxevent [args] -- command...\n")
		os.Exit(1)
	}

	cfg.command = os.Args[slicePosition("--", os.Args)+1:]

	run(cfg)
}

func run(cfg config) {
	start := time.Now()
	samples, err := executeCommand(cfg)
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

	if cfg.Influx.URL == "" || cfg.Influx.DB == "" {
		log.Printf("not writing point to influx since database or url is not set")
		os.Exit(exitStatus)
	}

	pt := point{
		measurement: cfg.Influx.Measurement,
		tags:        "etype=event",
		values:      map[string]float64{"duration": duration, "status": float64(exitStatus)},
		timestamp:   start,
	}

	samples = append(samples, pt)

	if cfg.Influx.DryRun {
		err = batchLogInfluxDB(dumpInfluxDB, cfg.Influx, samples)
	} else {
		err = batchLogInfluxDB(logInfluxDB, cfg.Influx, samples)
	}

	if err != nil {
		log.Printf("unable to write to influxdb: %v", err)
		os.Exit(1)
	}
}

func (p point) String() string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	influxString := fmt.Sprintf("%s,host=%s", p.measurement, host)
	if p.tags != "" {
		influxString = fmt.Sprintf("%s,%s", influxString, p.tags)
	}

	values := []string{}
	for k, v := range p.values {
		values = append(values, fmt.Sprintf("%s=%g", k, v))
	}
	influxString = fmt.Sprintf("%s %s %d", influxString, strings.Join(values, ","), p.timestamp.UnixNano())

	return influxString
}

func batchLogInfluxDB(fn func(influx, []byte) error, cfg influx, points []point) error {
	lines := []byte{}
	for i, pt := range points {
		pt.measurement = cfg.Measurement
		pt.tags += "," + cfg.Tags
		pt.tags = strings.Trim(pt.tags, ",")

		lines = append(lines, []byte(pt.String())...)
		lines = append(lines, '\n')
		if (i+1)%500 == 0 {
			err := fn(cfg, lines)
			if err != nil {
				return err
			}
			lines = []byte{}
		}
	}

	// Send remainder
	return fn(cfg, lines)
}

func dumpInfluxDB(cfg influx, lines []byte) error {
	var uri string
	if cfg.URL[len(cfg.URL)-1] == '/' {
		uri = fmt.Sprintf("%swrite?db=%s", cfg.URL, cfg.DB)
	} else {
		uri = fmt.Sprintf("%s/write?db=%s", cfg.URL, cfg.DB)
	}

	// Dangerous; shoud use url encoding
	if cfg.User != "" {
		uri += fmt.Sprintf("&u=%s&p=%s", cfg.User, cfg.Pass)
	}

	fmt.Println(uri)
	fmt.Println(string(lines))
	return nil
}

func logInfluxDB(cfg influx, lines []byte) error {
	var uri string
	if cfg.URL[len(cfg.URL)-1] == '/' {
		uri = fmt.Sprintf("%swrite?db=%s", cfg.URL, cfg.DB)
	} else {
		uri = fmt.Sprintf("%s/write?db=%s", cfg.URL, cfg.DB)
	}

	// Dangerous; shoud use url encoding
	if cfg.User != "" {
		uri += fmt.Sprintf("&u=%s&p=%s", cfg.User, cfg.Pass)
	}

	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Millisecond,
	}

	var err error
	var resp *http.Response

	r := bytes.NewReader(lines)

	for try := 0; try < cfg.Retries; try++ {
		resp, err = client.Post(uri, "application/x-www-form-urlencoded", r)
		if err == nil {
			break
		}
		fmt.Println(try)
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("unable to write to influxdb server %s, got response: %s", cfg.URL, resp.Status)
	}
	return nil
}

func executeCommand(cfg config) ([]point, error) {
	var cmd *exec.Cmd
	var ctx context.Context
	var cancel context.CancelFunc

	if cfg.Timeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, cfg.command[0], cfg.command[1:]...)
	} else {
		cmd = exec.Command(cfg.command[0], cfg.command[1:]...)
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
	scanErr := bufio.NewScanner(serr)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for scanOut.Scan() {
			fmt.Fprintf(os.Stdout, "%s\n", scanOut.Text())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for scanErr.Scan() {
			fmt.Fprintf(os.Stderr, "%s\n", scanErr.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	if cfg.Verbose {
		log.Printf("started subprocess %d\n", cmd.Process.Pid)
	}

	samplec := make(chan point, 500)
	var samples []point

	// Monitor process is a period is specified
	if cfg.Period > 0 {
		wg.Add(1)
		go watch(cmd.Process.Pid, cfg.Period, samplec, &wg)

		for s := range samplec {
			s.tags = "etype=metric," + s.tags
			s.tags = strings.Trim(s.tags, ",")
			samples = append(samples, s)
		}
	}

	wg.Wait()

	return samples, cmd.Wait()
}

func watch(pid int, period float64, c chan point, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(c)

	ticker := time.NewTicker(time.Duration(period) * time.Millisecond)

	for {
		t := <-ticker.C
		sysInfo, err := pidusage.GetStat(pid)
		if err != nil || sysInfo.Memory == 0 {

			return
		}
		c <- point{
			values: map[string]float64{
				"cpu":    sysInfo.CPU,
				"memory": sysInfo.Memory,
			},
			timestamp: t,
		}
	}
}
