package main

import (
	"reflect"
	"testing"
	"time"
)

func TestStrHash(t *testing.T) {
	// Test strings generated on the command line using
	// echo -n "<test string>" | sha256sum
	type args struct {
		s1 string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Good01", args{"Hello"}, "185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969"},
		{"Good02", args{"Another String"}, "8a43e989c31005cb19a8e61c513f352402b106b1fda0868f21f3d5708ae3d1a9"},
		{"Good03", args{"!@£"}, "2bf6f0ec863fa998b2f2cbb4eb9f6cb774d574c77d4c5d368d11710dd380a532"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrHash(tt.args.s1); got != tt.want {
				t.Errorf("StrHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newBackoff(t *testing.T) {
	type args struct {
		initial     time.Duration
		maxInterval time.Duration
		multiplier  float64
		maxRetries  int
	}
	tests := []struct {
		name string
		args args
		want *backoff
	}{
		{"Good01", args{10, 200, 2, 10}, &backoff{time.Second * 10, time.Second * 10, time.Second * 200, 2, 10, 0, 10}},
		{"Good02", args{1, 500, 10, 99}, &backoff{time.Second * 1, time.Second * 1, time.Second * 500, 10, 99, 0, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newBackoff(tt.args.initial, tt.args.maxInterval, tt.args.multiplier, tt.args.maxRetries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_backoff_Backoff(t *testing.T) {
	ch := make(chan error)
	type fields struct {
		initial      time.Duration
		current      time.Duration
		maxInterval  time.Duration
		multiplier   float64
		maxRetries   int
		currentRetry int
		jitter       int
	}
	type args struct {
		ec chan<- error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Good01", fields{1, 1, 1, 2, 10, 0, 4}, args{ch}, false},
		{"RetryFull", fields{1, 1, 1, 2, 10, 10, 4}, args{ch}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &backoff{
				initial:      tt.fields.initial,
				current:      tt.fields.current,
				maxInterval:  tt.fields.maxInterval,
				multiplier:   tt.fields.multiplier,
				maxRetries:   tt.fields.maxRetries,
				currentRetry: tt.fields.currentRetry,
				jitter:       tt.fields.jitter,
			}
			go b.Backoff(tt.args.ec)
			select {
			case err := <-ch:
				if tt.wantErr && err == nil {
					t.Errorf("Test_backoff_Backoff() = expected error but got nil")
				}
				if !tt.wantErr && err != nil {
					t.Errorf("Test_backoff_Backoff() = did not expected error but got: %v", err.Error())

				}
			case <-time.After(time.Second):
				t.Errorf("Test timed out")
			}
		})
	}
}
