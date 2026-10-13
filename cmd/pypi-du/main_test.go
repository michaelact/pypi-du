package main

import (
	"reflect"
	"testing"
)

func TestParseArgsInterspersed(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		wantProjects []string
		wantPerVer   bool
		wantUser     string
	}{
		{
			name:         "flag before positional",
			args:         []string{"--versions", "demo"},
			wantProjects: []string{"demo"},
			wantPerVer:   true,
		},
		{
			name:         "flag AFTER positional (the bug)",
			args:         []string{"demo", "--versions"},
			wantProjects: []string{"demo"},
			wantPerVer:   true,
		},
		{
			name:         "flag interspersed between positionals",
			args:         []string{"a", "--versions", "b"},
			wantProjects: []string{"a", "b"},
			wantPerVer:   true,
		},
		{
			name:         "multiple positionals no flags",
			args:         []string{"a", "b", "c"},
			wantProjects: []string{"a", "b", "c"},
		},
		{
			name:     "user flag after nothing",
			args:     []string{"--user", "pypa"},
			wantUser: "pypa",
		},
		{
			name:         "value flag after positional",
			args:         []string{"demo", "--limit", "8GB"},
			wantProjects: []string{"demo"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opt, err := parseArgs("pypi-du", c.args)
			if err != nil {
				t.Fatalf("parseArgs(%v): %v", c.args, err)
			}
			if c.wantProjects != nil && !reflect.DeepEqual(opt.projects, c.wantProjects) {
				t.Errorf("projects = %v, want %v", opt.projects, c.wantProjects)
			}
			if opt.perVer != c.wantPerVer {
				t.Errorf("perVer = %v, want %v", opt.perVer, c.wantPerVer)
			}
			if opt.user != c.wantUser {
				t.Errorf("user = %q, want %q", opt.user, c.wantUser)
			}
		})
	}
}

func TestParseArgsLimitValueAfterPositional(t *testing.T) {
	opt, err := parseArgs("pypi-du", []string{"demo", "--limit", "8GB"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if opt.limitStr != "8GB" {
		t.Errorf("limitStr = %q, want 8GB", opt.limitStr)
	}
}
