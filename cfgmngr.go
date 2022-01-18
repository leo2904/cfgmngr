package cfgmngr

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	flags "github.com/jessevdk/go-flags"
)

// version 1.0.5
// buildDate 2020-11-07

// AppName constant is used to indicate to the Parser function that the TOML's filename
// should be set after the application's name
const AppName = "APPNAME"

// Parse fills the received struct with the command line parameters
// and merges it with the parameters in the TOML file
// The TOMLFile parameter could be the real fullpath TOML file or the constant AppName in which case the
// Parse function will be looking for a TOML file named after the application's binary file
func Parse(cfg interface{}, TOMLFile string) (action string, err error) {
	action = getAction(&os.Args)
	if err := doParse(cfg, TOMLFile); err != nil {
		return "", err
	}
	return action, nil
}

// ParseWithVersion fills the received struct (calling Parse) and also gives support to display a standard version message
// The TOMLFile parameter could be the real fullpath TOML file or the constant AppName in which case the
// Parse function will be looking for a TOML file named after the application's binary file
// There is no manipulation in the version and date information so be sure to pass it in the format the user should see
func ParseWithVersion(cfg interface{}, TOMLFile string, appVersion string, appBuildDate string, versionFlag *bool) (action string, err error) {
	action = getAction(&os.Args)
	if err := doParseWithVersion(cfg, TOMLFile, appVersion, appBuildDate, versionFlag); err != nil {
		return "", err
	}
	return action, nil
}

// getAction returns the action, if exists, and
// strips it from the args slice
func getAction(args *[]string) string {
	if len(*args) < 2 {
		return ""
	}
	action := (*args)[1]

	if strings.HasPrefix(action, "-") || strings.HasPrefix(action, "/") {
		return ""
	}

	for i := 1; i < len(*args)-1; i++ {
		(*args)[i] = (*args)[i+1]
	}
	*args = (*args)[:len(*args)-1]

	return action
}

func doParseWithVersion(cfg interface{}, TOMLFile string, appVersion string, appBuildDate string, versionFlag *bool) error {
	if err := doParse(cfg, tomlFileName(TOMLFile)); err != nil {
		return err
	}

	if *versionFlag {
		fmt.Printf("%s version: %s [%s]\n", appName(), appVersion, appBuildDate)
		os.Exit(0)
	}

	return nil
}

func doParse(cfg interface{}, TOMLFile string) error {
	if _, err := flags.Parse(cfg); err != nil {
		if flagsErr, ok := err.(*flags.Error); !ok || flagsErr.Type != flags.ErrHelp {
			return err
		}
		os.Exit(0)
	}

	configFile := configFile(tomlFileName(TOMLFile))
	if configFile != "" {
		var tomlCfg map[string]interface{}
		if _, err := toml.DecodeFile(configFile, &tomlCfg); err != nil {
			return err
		}
		merge(cfg, tomlCfg)
	}

	return nil
}

func tomlFileName(TOMLFile string) string {
	if TOMLFile == AppName || TOMLFile == "" {
		return appName() + ".toml"
	}
	return TOMLFile
}

func appName() string {
	_, app := filepath.Split(os.Args[0])
	return strings.TrimSuffix(app, filepath.Ext(app))
}

func merge(config interface{}, tomlCfg map[string]interface{}) {
	cfg := reflect.ValueOf(config).Elem()

	for i := 0; i < cfg.NumField(); i++ {
		fv := cfg.Field(i)
		ft := cfg.Type().Field(i)

		if !fv.CanSet() {
			continue
		}

		tag := ft.Tag.Get("toml")
		if tag == "" || tag == "-" {
			continue
		}

		tomlValue, ok := tomlCfg[tag]
		if !ok || tomlValue == "-" {
			continue
		}

		cfgValue := cfg.Field(i).Interface()

		switch cfg.Field(i).Kind() {
		case reflect.String:
			if cfgValue != "" {
				continue
			}

		case reflect.Int:
			if cfgValue != 0 {
				continue
			}

		case reflect.Bool:
			if cfgValue == true {
				continue
			}

		case reflect.Map:
			if fv.Len() > 0 {
				continue
			}

		case reflect.Slice:
			if fv.Len() > 0 {
				continue
			}
		}

		setValue(fv, tomlValue)
	}
}

func setValue(fv reflect.Value, v interface{}) {
	switch fv.Kind() {
	case reflect.String:
		convertAndSet(v.(string), fv)

	case reflect.Int:
		fv.SetInt(v.(int64))

	case reflect.Float64:
		fv.SetFloat(v.(float64))

	case reflect.Bool:
		fv.SetBool(v.(bool))

	case reflect.Uint64:
		fv.SetUint(v.(uint64))

	case reflect.Map:
		iv := v.([]interface{})
		for i := 0; i < len(iv); i++ {
			addToMap(fv, iv[i])
		}

	case reflect.Slice:
		iv := v.([]interface{})
		sv := make([]string, len(iv))
		for i := 0; i < len(iv); i++ {
			sv[i] = fmt.Sprint(iv[i])
		}
		fv.Set(reflect.AppendSlice(fv, reflect.ValueOf(sv)))
	}
}

func addToMap(fv reflect.Value, pair interface{}) {
	av := pair.([]interface{})
	key := av[0].(string)
	var value string

	if len(av) == 2 {
		value = av[1].(string)
	}

	convertAndSet(fmt.Sprintf("%s:%s", key, value), fv)
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

func convertAndSet(val string, retval reflect.Value) error {
	tp := retval.Type()

	// Support for time.Duration
	if tp == reflect.TypeOf((*time.Duration)(nil)).Elem() {
		parsed, err := time.ParseDuration(val)

		if err != nil {
			return err
		}

		retval.SetInt(int64(parsed))
		return nil
	}

	switch tp.Kind() {
	case reflect.String:
		retval.SetString(val)
	case reflect.Bool:
		if val == "" {
			retval.SetBool(true)
		} else {
			b, err := strconv.ParseBool(val)

			if err != nil {
				return err
			}

			retval.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(val, 10, tp.Bits())
		if err != nil {
			return err
		}

		retval.SetInt(parsed)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(val, 10, tp.Bits())
		if err != nil {
			return err
		}

		retval.SetUint(parsed)

	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(val, tp.Bits())
		if err != nil {
			return err
		}

		retval.SetFloat(parsed)

	case reflect.Slice:
		elemtp := tp.Elem()

		elemvalptr := reflect.New(elemtp)
		elemval := reflect.Indirect(elemvalptr)
		if err := convertAndSet(val, elemval); err != nil {
			return err
		}

		retval.Set(reflect.Append(retval, elemval))

	case reflect.Map:
		parts := strings.SplitN(val, ":", 2)

		key := parts[0]
		var value string

		if len(parts) == 2 {
			value = parts[1]
		}

		keytp := tp.Key()
		keyval := reflect.New(keytp)

		if err := convertAndSet(key, keyval); err != nil {
			return err
		}

		valuetp := tp.Elem()
		valueval := reflect.New(valuetp)

		if err := convertAndSet(value, valueval); err != nil {
			return err
		}

		if retval.IsNil() {
			retval.Set(reflect.MakeMap(tp))
		}

		retval.SetMapIndex(reflect.Indirect(keyval), reflect.Indirect(valueval))

	case reflect.Ptr:
		if retval.IsNil() {
			retval.Set(reflect.New(retval.Type().Elem()))
		}

		return convertAndSet(val, reflect.Indirect(retval))

	case reflect.Interface:
		if !retval.IsNil() {
			return convertAndSet(val, retval.Elem())
		}
	}
	return nil
}
