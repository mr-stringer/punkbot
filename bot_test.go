package main

import (
	"testing"
)

func TestCheckForTerms(t *testing.T) {
	tp1 := &Message{Commit: Commit{Record: Record{Text: "I ran 10K #runningpunks"}}}
	tp2 := &Message{Commit: Commit{Record: Record{Text: "I love bread, it's nice"}}}
	tp3 := &Message{Commit: Commit{Record: Record{Text: "I love breads, they're nice"}}}
	tp4 := &Message{Commit: Commit{Record: Record{Text: ""}}}
	type args struct {
		cnf *Config
		msg *Message
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"GoodMatch1", args{&Config{Terms: []string{"#runningpunks"}}, tp1}, true},
		{"GoodMatch2", args{&Config{Terms: []string{"bread"}}, tp2}, true},
		{"GoodNoMatch1", args{&Config{Terms: []string{"#robots"}}, tp1}, false},
		{"GoodNoMatch2", args{&Config{Terms: []string{"#runningpunksarebad"}}, tp1}, false},
		{"GoodNoMatch3", args{&Config{Terms: []string{"nicebread"}}, tp2}, false},
		{"GoodNoMatch3", args{&Config{Terms: []string{"bread"}}, tp3}, false},
		{"NoText", args{&Config{Terms: []string{"bread"}}, tp4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkForTerms(tt.args.cnf, tt.args.msg); got != tt.want {
				t.Errorf("checkForTerms() = %v, want %v", got, tt.want)
			}
		})
	}
}
