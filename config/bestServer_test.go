package config

import (
	"fmt"
	"testing"
	"time"
)

var MockLatencyTestGoodCounter time.Duration = 0

// This mock latency test uses a counter. It increments by 100 and then returns
func MockLatencyTestGood(address string) (*time.Duration, error) {
	MockLatencyTestGoodCounter += 100

	return &MockLatencyTestGoodCounter, nil
}

// This mock latency test uses a always returns an error
func MockLatencyTestErr(address string) (*time.Duration, error) {
	return nil, fmt.Errorf("there was a problem")
}
func TestConfig_FindFastestServer(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Good01", fields{"someDID", []string{"punk", "punkrock"}, "", ""}, args{MockLatencyTestGood}, false},
		{"Error", fields{"someDID", []string{"punk", "punkrock"}, "", ""}, args{MockLatencyTestErr}, true},
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
		})
	}
}
