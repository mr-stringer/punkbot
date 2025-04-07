package config

import (
	"os"
	"testing"
)

func TestConfig_GetSecretFromEnv(t *testing.T) {
	type fields struct {
		Identifier string
		Hashtags   []string
		password   string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"Good01", fields{"punkbot", []string{"punkbot", "running"}, "asd-fgh-jkl"}, false},
		{"BadNoPassword", fields{"punkbot", []string{"punkbot", "running"}, ""}, true},
	}
	for _, tt := range tests {
		switch tt.name {
		case "Good01":
			os.Setenv("PUNKBOT_PASSWORD", "asd-fgh-jkl")
		case "BadNoPassword":
			os.Setenv("PUNKBOT_PASSWORD", "")
		}
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Identifier: tt.fields.Identifier,
				Terms:      tt.fields.Hashtags,
				password:   tt.fields.password,
			}
			if err := c.GetSecretFromEnv(); (err != nil) != tt.wantErr {
				t.Errorf("Config.GetSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetSecret(t *testing.T) {
	type fields struct {
		Identifier      string
		Hashtags        []string
		JetStreamServer string
		password        string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Good01", fields{"MyBot", []string{"tag1", "tag2"}, "", "NotGoodPassword"}, "NotGoodPassword"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Identifier:      tt.fields.Identifier,
				Terms:           tt.fields.Hashtags,
				JetStreamServer: tt.fields.JetStreamServer,
				password:        tt.fields.password,
			}
			if got := c.GetSecret(); got != tt.want {
				t.Errorf("Config.GetSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
