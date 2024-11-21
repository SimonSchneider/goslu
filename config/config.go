package config

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type configField struct {
	field reflect.Value
	dt    reflect.Kind
	path  string
	name  string
	usage string

	boolVal bool
	strVal  string
	intVal  int64
	uintVal uint64
}

func validKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		return true
	default:
		return false
	}
}

func parseTag(tag string) (string, bool) {
	if strings.Contains(tag, "omit") {
		return "", true
	}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "u:") {
			return part[2:], false
		}
	}
	return "", false
}

func appendStructFields(fields []configField, val reflect.Value, path string) ([]configField, error) {
	if !val.IsValid() {
		return nil, fmt.Errorf("invalid value: %s", val.String())
	}
	for i := 0; i < 10; i++ {
		if (val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface) && !val.IsNil() {
			val = val.Elem()
		}
	}
	styp := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fld := val.Field(i)
		typ := styp.Field(i)
		usage, omit := parseTag(typ.Tag.Get("config"))
		pth := path + typ.Name
		if omit {
			// skip
		} else if fld.Kind() == reflect.Struct {
			var err error
			fields, err = appendStructFields(fields, fld, pth+".")
			if err != nil {
				return nil, err
			}
		} else if fld.IsValid() && fld.CanSet() && validKind(fld.Kind()) {
			fields = append(fields, configField{
				field: fld,
				path:  pth,
				dt:    fld.Kind(),
				usage: usage,
			})
		}
	}
	return fields, nil
}

func getFields(cfg any) ([]configField, error) {
	return appendStructFields(nil, reflect.ValueOf(cfg), "")
}

type Parser struct {
	parsed  bool
	FlagSet *flag.FlagSet
	Args    []string
	GetEnv  func(string) string
	FlagSep string
	EnvSep  string
}

func parseBool(val string) bool {
	switch val {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	default:
		return false
	}
}

func parseInt(val string) int64 {
	if val == "" {
		return 0
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func parseUint(val string) uint64 {
	if val == "" {
		return 0
	}
	i, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

type FlagVal struct {
	name string
	cb   func(string) error
}

func NewVal(name string, cb func(string) error) FlagVal {
	return FlagVal{name: name, cb: cb}
}

func (f FlagVal) String() string {
	return f.name
}

func (f FlagVal) Set(s string) error {
	return f.cb(s)
}

func (p *Parser) ParseInto(cfg any) error {
	if p.parsed {
		panic("already parsed")
	}
	p.parsed = true
	fields, err := getFields(cfg)
	if err != nil {
		return err
	}
	for i, fld := range fields {
		flagName := strings.ToLower(strings.ReplaceAll(fld.path, ".", p.FlagSep))
		if flagName == "" {
			panic(fmt.Sprintf("illegal empty flag name: %s", flagName))
		}
		envName := strings.ToUpper(strings.ReplaceAll(fld.path, ".", p.EnvSep))
		if envName == "" {
			panic(fmt.Sprintf("illegal empty env name: %s", envName))
		}
		envVal := p.GetEnv(envName)
		switch fld.dt {
		case reflect.Bool:
			fld.field.SetBool(parseBool(envVal))
			p.FlagSet.BoolFunc(flagName, fld.usage, func(val string) error {
				fld.field.SetBool(parseBool(val))
				return nil
			})
		case reflect.String:
			p.FlagSet.StringVar(&fields[i].strVal, flagName, fld.field.String(), fld.usage)
			fld.field.SetString(envVal)
		case reflect.Int, reflect.Int64:
			p.FlagSet.Int64Var(&fields[i].intVal, flagName, fld.field.Int(), fld.usage)
			fld.field.SetInt(parseInt(envVal))
		case reflect.Uint, reflect.Uint64:
			p.FlagSet.Uint64Var(&fields[i].uintVal, flagName, fld.field.Uint(), fld.usage)
			fld.field.SetUint(parseUint(envVal))
		default:
			panic(fmt.Sprintf("unsupported kind: %s", fld.dt))
		}
	}
	if err := p.FlagSet.Parse(p.Args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	for _, fld := range fields {
		if fld.dt == reflect.Bool {
			//fld.field.SetBool(fld.boolVal)
		} else if fld.dt == reflect.String {
			fld.field.SetString(fld.strVal)
		} else if fld.dt == reflect.Int || fld.dt == reflect.Int64 {
			fld.field.SetInt(fld.intVal)
		} else if fld.dt == reflect.Uint || fld.dt == reflect.Uint64 {
			fld.field.SetUint(fld.uintVal)
		}
	}
	return nil
}

func ParseInto(cfg any, flagSet *flag.FlagSet, args []string, getEnv func(string) string) error {
	parser := &Parser{
		FlagSet: flagSet,
		GetEnv:  getEnv,
		Args:    args,
		FlagSep: "-",
		EnvSep:  "_",
	}
	return parser.ParseInto(&cfg)
}
