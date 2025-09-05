package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var MockLatencyTestGoodCounterAsc time.Duration = 0
var MockLatencyTestGoodCounterDsc time.Duration = 1000

// This mock latency test uses a counter. It increments by 100 and then returns
func MockLatencyTestGoodAsc(address string) (*time.Duration, error) {
	MockLatencyTestGoodCounterAsc += 100

	return &MockLatencyTestGoodCounterAsc, nil
}

// This mock latency test uses a counter. It decrements by 100 and then returns
func MockLatencyTestGoodDsc(address string) (*time.Duration, error) {
	MockLatencyTestGoodCounterDsc -= 100

	return &MockLatencyTestGoodCounterDsc, nil
}

// This mock latency test uses a always returns an error
func MockLatencyTestErr(address string) (*time.Duration, error) {
	return nil, fmt.Errorf("there was a problem")
}
func TestFindFastestServer(t *testing.T) {
	stderr := os.Stderr
	defer func() { os.Stdout = stderr }()
	os.Stderr = os.NewFile(0, os.DevNull)
	type fields struct {
		Identifier      string
		Hashtags        []string
		JetStreamServer string
		password        string
	}
	type args struct {
		ml MeasureLatencyType
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantServer string
	}{
		{"Good01", fields{"someDID", []string{"punk", "punkrock"}, "", ""}, args{MockLatencyTestGoodAsc}, false, "jetstream1.us-east.bsky.network"},
		{"Good02", fields{"someDID", []string{"punk", "punkrock"}, "", ""}, args{MockLatencyTestGoodDsc}, false, "jetstream2.us-west.bsky.network"},
		{"Error", fields{"someDID", []string{"punk", "punkrock"}, "", ""}, args{MockLatencyTestErr}, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Identifier:      tt.fields.Identifier,
				Terms:           tt.fields.Hashtags,
				JetStreamServer: tt.fields.JetStreamServer,
				password:        tt.fields.password,
			}
			if err := c.FindFastestServer(tt.args.ml); (err != nil) != tt.wantErr {
				t.Errorf("Config.FindFastestServer() error = %v, wantErr %v", err, tt.wantErr)
			}
			/* if no error is wanted, also check jetstream server is as expected */
			if !tt.wantErr {
				if tt.wantServer != c.JetStreamServer {
					t.Errorf("Config.FindFastestServer() expected jetstream server to nbe %s, but found %s", tt.wantServer, c.JetStreamServer)
				}
			}
		})
	}
}
