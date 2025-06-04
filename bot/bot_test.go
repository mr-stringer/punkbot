package bot

import (
	"testing"

	"github.com/mr-stringer/punkbot/config"
	"github.com/mr-stringer/punkbot/global"
)

func Test_checkForTerms(t *testing.T) {
	tp1 := &global.Message{Commit: global.Commit{Record: global.Record{Text: "I ran 10K #runningpunks"}}}
	tp2 := &global.Message{Commit: global.Commit{Record: global.Record{Text: "I love bread, it's nice"}}}
	tp3 := &global.Message{Commit: global.Commit{Record: global.Record{Text: "I love breads, they're nice"}}}
	type args struct {
		cnf *config.Config
		msg *global.Message
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"GoodMatch1", args{&config.Config{Terms: []string{"#runningpunks"}}, tp1}, true},
		{"GoodMatch2", args{&config.Config{Terms: []string{"bread"}}, tp2}, true},
		{"GoodNoMatch1", args{&config.Config{Terms: []string{"#robots"}}, tp1}, false},
		{"GoodNoMatch2", args{&config.Config{Terms: []string{"#runningpunksarebad"}}, tp1}, false},
		{"GoodNoMatch3", args{&config.Config{Terms: []string{"nicebread"}}, tp2}, false},
		{"GoodNoMatch3", args{&config.Config{Terms: []string{"bread"}}, tp3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkForTerms(tt.args.cnf, tt.args.msg); got != tt.want {
				t.Errorf("checkForTerms() = %v, want %v", got, tt.want)
			}
		})
	}
}
