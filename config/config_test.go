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
	cfg := Cfg{
		Host:    "otherhost",
		ignored: "hello",
	}
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
	failIfNot(t, cfg.Host != "localhost", "host")
	failIfNot(t, cfg.Name != "name", "name")
	failIfNot(t, cfg.Enabled != true, "enabled")
	failIfNot(t, cfg.DoAgain != true, "doagain")
	failIfNot(t, cfg.Retries != 43, "retries")
	failIfNot(t, cfg.Extra.Foo != "foo", "extra.foo")
	failIfNot(t, cfg.Extra.Bar != 42, "extra.bar")
	failIfNot(t, cfg.ignored != "hello", "ignored")
}

func failIfNot(t *testing.T, b bool, str string) {
	t.Helper()
	if b {
		t.Error(str)
	}
}
