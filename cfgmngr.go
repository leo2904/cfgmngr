package cfgmngr

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
	flags "github.com/jessevdk/go-flags"
)

// // ParseWithVersion does whatever Parse does and gives support to display version information
// func ParseWithVersion(cfg interface{}, TOMLFile string, buildVersion string, buildDate string) error {
// 	if hasField(cfg,"Version") && cfg.Version == true {
// 		_, app := filepath.Split(os.Args[0])
// 		app = strings.TrimSuffix(app, filepath.Ext(app))
// 		fmt.Printf("%s version: %s [%s]\n", app, buildVersion, buildDate)
// 		os.Exit(0)
// 	}

// 	return nil
// }

// Parse fills the received struct with the command line parameters
// and merges it with the parameters in the TOML file
func Parse(cfg interface{}, TOMLFile string) error {
	if _, err := flags.Parse(cfg); err != nil {
		if flagsErr, ok := err.(*flags.Error); !ok || flagsErr.Type != flags.ErrHelp {
			return err
		}
		os.Exit(0)
	}

	configFile := configFile(TOMLFile)

	// if TOMLFile != "" {
	if configFile != "" {
		var tomlCfg map[string]interface{}
		// if _, err := os.Stat(TOMLFile); !os.IsNotExist(err) {
		// if _, err := toml.DecodeFile(TOMLFile, &tomlCfg); err != nil {
		// 	return err
		// }
		// }
		if _, err := toml.DecodeFile(configFile, &tomlCfg); err != nil {
			return err
		}
		merge(cfg, tomlCfg)
	}

	return nil
}

func configFile(TOMLFile string) string {
	if TOMLFile != "" {
		dir, _ := filepath.Split(TOMLFile)
		if dir != "" {
			if _, err := os.Stat(TOMLFile); os.IsNotExist(err) {
				return ""
			}
			return TOMLFile
		}

		// is not a qualified path so I will try to find the config file

		// first of all I look into the current directory
		if _, err := os.Stat(TOMLFile); !os.IsNotExist(err) {
			return TOMLFile
		}

		// if it is not there I check the exe directory
		exeDir, _ := filepath.Split(os.Args[0])
		if _, err := os.Stat(filepath.Join(exeDir, TOMLFile)); !os.IsNotExist(err) {
			return filepath.Join(exeDir, TOMLFile)
		}

		// if still not there then the last resort is the CONFIG_REPO env variable
		envDir := os.Getenv("CONFIG_REPO")
		if _, err := os.Stat(filepath.Join(envDir, TOMLFile)); !os.IsNotExist(err) {
			return filepath.Join(envDir, TOMLFile)
		}
	}
	return ""
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

func hasField(Iface interface{}, FieldName string) bool {
	ValueIface := reflect.ValueOf(Iface)

	// Check if the passed interface is a pointer
	if ValueIface.Type().Kind() != reflect.Ptr {
		// Create a new type of Iface's Type, so we have a pointer to work with
		ValueIface = reflect.New(reflect.TypeOf(Iface))
	}

	// 'dereference' with Elem() and get the field by name
	Field := ValueIface.Elem().FieldByName(FieldName)
	if !Field.IsValid() {
		return false
	}
	return true
}
