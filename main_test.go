package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

func Test_point_String(t *testing.T) {
	host, _ := os.Hostname()

	type fields struct {
		measurement string
		tags        string
		duration    float64
		status      int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{

			name: "field 1",
			fields: fields{
				measurement: "foo",
				tags:        "bar=baz",
				duration:    1.234,
				status:      255,
			},
			want: "foo,host=" + host + ",bar=baz duration=1.234,status=255",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := point{
				measurement: tt.fields.measurement,
				tags:        tt.fields.tags,
				duration:    tt.fields.duration,
				status:      tt.fields.status,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("point.String() = %v, want %v", got, tt.want)
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
					db:   "foodb",
					user: "u", pass: "p",
				},
				point: point{
					measurement: "events",
					tags:        "foo=bar",
					duration:    9.876,
					status:      0,
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
		tt.args.server.url = inf.URL
		t.Run(tt.name, func(t *testing.T) {
			if err := logInfluxDB(tt.args.server, tt.args.point); (err != nil) != tt.wantErr {
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
			args:    args{args: []string{"ls", "/tmp"}},
			wantErr: false,
		},
		{
			name: "ls tmp",
			args: args{
				args:    []string{"sleep", "0.1"},
				timeout: 0.1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := executeCommand(tt.args.args, tt.args.timeout); (err != nil) != tt.wantErr {
				t.Errorf("executeCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
