package cfgmngr

import (
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	flags "github.com/jessevdk/go-flags"
)

// Parse fills the received struct with the command line parameters
// and merges it with the parameters in the TOML file
func Parse(cfg interface{}, TOMLFile string) error {
	if _, err := flags.Parse(cfg); err != nil {
		if flagsErr, ok := err.(*flags.Error); !ok || flagsErr.Type != flags.ErrHelp {
			return err
		}
		os.Exit(0)
	}

	var tomlCfg map[string]interface{}

	if TOMLFile != "" {
		if _, err := os.Stat(TOMLFile); !os.IsNotExist(err) {
			if _, err := toml.DecodeFile(TOMLFile, &tomlCfg); err != nil {
				return err
			}
		}
		merge(cfg, tomlCfg)
	}
	return nil
}

func merge(cfg interface{}, tomlCfg map[string]interface{}) {
	s := reflect.ValueOf(cfg).Elem()
	for i := 0; i < s.NumField(); i++ {
		fv := s.Field(i)
		ft := s.Type().Field(i)

		if !fv.CanSet() {
			continue
		}

		if fv.String() == "" {
			v, ok := tomlCfg[ft.Tag.Get("toml")]
			if !ok || v == "-" {
				def := ft.Tag.Get("def")
				setValue(fv, def)
				continue
			}
			setValue(fv, v)
		}
	}
}

func setValue(fv reflect.Value, v interface{}) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(v.(string))
	case reflect.Int:
		fv.SetInt(v.(int64))
	case reflect.Float64:
		fv.SetFloat(v.(float64))
	case reflect.Bool:
		fv.SetBool(true)
	case reflect.Uint64:
		fv.SetUint(v.(uint64))
	}
}
