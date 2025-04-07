package config

import (
	"os"
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{"Good01", args{"../testfiles/Good01.yml"}, &Config{"punkbot", []string{"punkbot", "running"}, "my.jetstream.example.com", "SomePassword"}, false},
		{"Good02", args{"../testfiles/Good02.json"}, &Config{"punkbotjson", []string{"punkbot", "running", "JSON"}, "my.jetstream.example.com", "SomePassword"}, false},
		{"Good03", args{"../testfiles/Good03.yml"}, &Config{"punkbot", []string{"punkbot", "running"}, "my.jetstream.example.com", "SomePassword"}, false},
		{"MalformedJson", args{"../testfiles/Malformed.json"}, nil, true},
		{"NoFile", args{"../testfiles/NoFile.yml"}, nil, true},
		{"NoPassword", args{"../testfiles/NoPassword.yml"}, nil, true},
	}
	for _, tt := range tests {
		if tt.name == "NoPassword" {
			os.Unsetenv("PUNKBOT_PASSWORD")
		} else {
			os.Setenv("PUNKBOT_PASSWORD", "SomePassword")
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				//debug line
				t.Errorf("%s", err.Error())
				//end debug
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
