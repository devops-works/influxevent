package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_slicePosition(t *testing.T) {
	tests := []struct {
		needle string
		stack  []string
		pos    int
	}{
		{"foo", []string{"foo", "bar", "baz"}, 0},
		{"baz", []string{"foo", "bar", "baz"}, 2},
		{"fizz", []string{"foo", "bar", "baz"}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.needle, func(t *testing.T) {
			if slicePosition(tt.needle, tt.stack) != tt.pos {
				t.Errorf("needle %s not found at position %d in %s", tt.needle, tt.pos, tt.stack)
			}
		})
	}
}

func Test_anyInSlice(t *testing.T) {
	tests := []struct {
		needle []string
		stack  []string
		in     bool
	}{
		{[]string{"foo", "fizz"}, []string{"foo", "bar", "baz"}, true},
		{[]string{"bar", "baz", "buzz"}, []string{"foo", "bar", "baz"}, true},
		{[]string{"bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{[]string{"fizz", "buzz"}, []string{"foo", "bar", "baz"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.needle[0], func(t *testing.T) {
			if anyInSlice(tt.needle, tt.stack) != tt.in {
				t.Errorf("any of %s yields %t found in %s", tt.needle, tt.in, tt.stack)
			}
		})
	}
}

func Test_point_String(t *testing.T) {
	host, _ := os.Hostname()
	now := time.Now()

	type fields struct {
		measurement string
		tags        string
		values      map[string]float64
		timestamp   time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{

			name: "field 1",
			fields: fields{
				measurement: "foo",
				tags:        "bar=baz",
				values:      map[string]float64{"duration": 1.234, "status": 255},
				timestamp:   now,
			},
			want: []string{
				fmt.Sprintf("foo,host=%s,bar=baz duration=1.234,status=255 %d", host, now.UnixNano()),
				fmt.Sprintf("foo,host=%s,bar=baz status=255,duration=1.234 %d", host, now.UnixNano()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := point{
				measurement: tt.fields.measurement,
				tags:        tt.fields.tags,
				values:      tt.fields.values,
				timestamp:   tt.fields.timestamp,
			}
			got := p.String()
			ok := false
			for _, s := range tt.want {
				if s == got {
					ok = true
				}
			}
			if !ok {
				t.Errorf("point.String() = %v, want one of %#v", got, tt.want)
			}
		})
	}
}

func Test_logInfluxDB(t *testing.T) {
	type args struct {
		server influx
		point  point
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1",
			args: args{
				server: influx{
					DB:   "foodb",
					User: "u", Pass: "p",
					Retries: 1,
				},
				point: point{
					measurement: "events",
					tags:        "foo=bar",
					values:      map[string]float64{"duration": 9.876, "status": 0},
				},
			},
			wantErr: false,
		},
	}

	inf := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, "OK")
	}))
	defer inf.Close()

	for _, tt := range tests {
		tt.args.server.URL = inf.URL
		t.Run(tt.name, func(t *testing.T) {
			if err := logInfluxDB(tt.args.server, []byte(tt.args.point.String())); (err != nil) != tt.wantErr {
				t.Errorf("logInfluxDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_executeCommand(t *testing.T) {
	type args struct {
		args    []string
		timeout float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "ls nowhere",
			args:    args{args: []string{"ls", "/nowhere"}},
			wantErr: true,
		},
		{
			name:    "ls tmp",
			args:    args{args: []string{"ls", "/"}},
			wantErr: false,
		},
		{
			name: "sleep too long",
			args: args{
				args:    []string{"sleep", "0.2"},
				timeout: 0.1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config{command: tt.args.args, Timeout: tt.args.timeout}
			if _, err := executeCommand(cfg); (err != nil) != tt.wantErr {
				t.Errorf("executeCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
