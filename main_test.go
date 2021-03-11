package main

import "testing"

func Test_injectInt(t *testing.T) {
	type args struct {
		filename string
		i        int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"nodot", args{"foo", 1}, "foo-1"},
		{"onedot", args{"foo.jpeg", 10}, "foo-10.jpeg"},
		{"unixhidden", args{".DS_Store", 0}, ".DS_Store-0"},
		{"unixhidden-withext", args{".config.js", 32}, ".config-32.js"},
		{"trailingdot", args{"config.js.", 1}, "config.js-1."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := injectInt(tt.args.filename, tt.args.i); got != tt.want {
				t.Errorf("injectInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
