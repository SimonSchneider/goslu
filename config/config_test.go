package config_test

import (
	"flag"
	"github.com/SimonSchneider/goslu/config"
	"testing"
)

func TestParseInto(t *testing.T) {
	type Cfg struct {
		Host    string
		Name    string
		Enabled bool
		DoAgain bool
		Retries int
		Extra   struct {
			Foo string
			Bar int
		}
		ignored string
	}
	var cfg Cfg
	fset := flag.NewFlagSet("", flag.ExitOnError)
	envs := map[string]string{
		"NAME":      "name",
		"RETRIES":   "43",
		"DOAGAIN":   "true",
		"HOST":      "",
		"EXTRA_FOO": "foo-from-env",
	}
	if err := config.ParseInto(&cfg, fset, []string{
		"-host", "localhost",
		"-extra-foo", "foo",
		"-enabled",
		"-extra-bar", "42",
	}, func(k string) string { return envs[k] }); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", cfg)
}

func TestParseWithDefaults(t *testing.T) {
	type Cfg struct {
		Foo string
		Bar int
	}
	cfg := Cfg{
		Foo: "foo",
		Bar: 42,
	}
	fset := flag.NewFlagSet("", flag.ExitOnError)
	if err := config.ParseInto(&cfg, fset, []string{}, func(k string) string { return "" }); err != nil {
		t.Error(err)
	}
	t.Logf("%+v", cfg)
}
