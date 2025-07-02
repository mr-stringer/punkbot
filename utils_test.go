package main

import "testing"

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
