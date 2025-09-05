package main

import (
	"testing"
)

func TestConfig_setAutoJetStream(t *testing.T) {
	type fields struct {
		Identifier      string
		Terms           []string
		JetStreamServer string
		autoJetStream   bool
		password        string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"Good01", fields{"username.bsky.app", []string{"news"}, "", false, "blah-blah-blah-blah"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Identifier:      tt.fields.Identifier,
				Terms:           tt.fields.Terms,
				JetStreamServer: tt.fields.JetStreamServer,
				autoJetStream:   tt.fields.autoJetStream,
				password:        tt.fields.password,
			}
			c.setAutoJetStream()
			if c.JetStreamServer == "" {
				if !c.autoJetStream {
					t.Errorf("Expected autoJetStream to be true, but it is set to false ")
				}
			}
		})
	}
}

func TestConfig_getAutoJetStream(t *testing.T) {
	type fields struct {
		Identifier      string
		Terms           []string
		JetStreamServer string
		autoJetStream   bool
		password        string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"AutoJetStream", fields{"username.bsky.app", []string{"news"}, "", true, "blah-blah-blah-blah"}, true},
		{"AutoJetStream", fields{"username.bsky.app", []string{"news"}, "myPrivateJs.internal.net", false, "blah-blah-blah-blah"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Identifier:      tt.fields.Identifier,
				Terms:           tt.fields.Terms,
				JetStreamServer: tt.fields.JetStreamServer,
				autoJetStream:   tt.fields.autoJetStream,
				password:        tt.fields.password,
			}
			if got := c.getAutoJetStream(); got != tt.want {
				t.Errorf("Config.getAutoJetStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
