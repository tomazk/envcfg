/*

Un-marshaling environment variables to Go structs

Getting Started

Let's set a bunch of environment variables and then run your go app

	#!/usr/bin/env bash
	export DEBUG="false"
	export DB_HOST="localhost"
	export DB_PORT="8012"

	./your_go_app

Within your Go app do

	import "github.com/tomazk/envcfg"

	// declare a type that will hold your env variables
	type Cfg struct {
		DEBUG bool
		DB_PORT int
		DB_HOST string
	}

	func main() {
		var config Cfg
		envcfg.Unmarshal(&config)
		// config is now set to Config{DEBUG: false, DB_PORT: 8012, DB_HOST: "localhost"}

		// optional: clear env variables listed in the Cfg struct
		envcfg.ClearEnvVars(&config)

	}

More documentation in README: https://github.com/tomazk/envcfg
*/
package envcfg

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	structTag     = "envcfg"
	structTagKeep = "envcfgkeep"
)

// Unmarshal will read your environment variables and try to unmarshal them
// to the passed struct. It will return an error, if it recieves an unsupported
// non-struct type, if types of the fields are not supported or if it can't
// parse value from an environment variable, thus taking care of validation of
// environment variables values. If failOnUndefined is set true, then Unmarshal will
// return an error if a value contained in the struct is undefined in the environment.
func Unmarshal(v interface{}, failOnUndefined bool) error {
	structType, err := makeSureTypeIsSupported(v)
	if err != nil {
		return err
	}
	if err := makeSureStructFieldTypesAreSupported(structType); err != nil {
		return err
	}
	makeSureValueIsInitialized(v)

	env, err := newEnviron()
	if err != nil {
		return err
	}

	structVal := getStructValue(v)

	if err := unmarshalAllStructFields(structVal, env, failOnUndefined); err != nil {
		return err
	}

	return nil
}

// ClearEnvVars will clear all environment variables based on the struct
// field names or struct field tags. It will keep all those with
// envcfgkeep:"" struct field tag. It will return an error,
// if it recieves an unsupported non-struct type, if types of the
// fields are not supported
func ClearEnvVars(v interface{}) error {
	structType, err := makeSureTypeIsSupported(v)
	if err != nil {
		return err
	}
	if err := makeSureStructFieldTypesAreSupported(structType); err != nil {
		return err
	}

	unsetEnvVars(structType)
	return nil
}

func unsetEnvVarFromSingleField(structField reflect.StructField) {
	if strings.Contains(string(structField.Tag), structTagKeep) {
		return
	}
	envKey := getEnvKey(structField)
	os.Setenv(envKey, "") // we're using Setenv instead of Unsetenv to ensure go1.3 compatibility
}

func unsetEnvVars(structType reflect.Type) {
	for i := 0; i < structType.NumField(); i++ {
		unsetEnvVarFromSingleField(structType.Field(i))
	}
}

func getEnvKey(structField reflect.StructField) string {
	if tag := structField.Tag.Get(structTag); tag != "" {
		return tag
	}
	return structField.Name
}

func undefinedInEnvironError(field string) error {
	return errors.New("field not found in environment: " + field)
}

func unmarshalInt(fieldVal reflect.Value, structField reflect.StructField, env environ, failOnUndefined bool) error {
	envKey := getEnvKey(structField)
	val, ok := env[envKey]
	if !ok {
		if failOnUndefined {
			return undefinedInEnvironError(envKey)
		} else {
			return nil
		}
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	fieldVal.SetInt(int64(i))
	return nil
}

var boolErr error = errors.New("pass string 'true' or 'false' for boolean fields")

func unmarshalBool(fieldVal reflect.Value, structField reflect.StructField, env environ, failOnUndefined bool) error {
	envKey := getEnvKey(structField)
	val, ok := env[envKey]
	if !ok {
		if failOnUndefined {
			return undefinedInEnvironError(envKey)
		} else {
			return nil
		}
	}

	var vbool bool
	switch val {
	case "true":
		vbool = true
	case "false":
		vbool = false
	default:
		return boolErr
	}

	fieldVal.SetBool(vbool)
	return nil
}

func unmarshalString(fieldVal reflect.Value, structField reflect.StructField, env environ, failOnUndefined bool) error {
	envKey := getEnvKey(structField)
	val, ok := env[envKey]
	if !ok {
		if failOnUndefined {
			return undefinedInEnvironError(envKey)
		} else {
			return nil
		}
	}

	fieldVal.SetString(val)
	return nil
}

func appendToStringSlice(fieldVal reflect.Value, sliceVal string) error {
	fieldVal.Set(reflect.Append(fieldVal, reflect.ValueOf(sliceVal)))
	return nil
}

func appendToIntSlice(fieldVal reflect.Value, sliceVal string) error {
	val, err := strconv.Atoi(sliceVal)
	if err != nil {
		return err
	}
	fieldVal.Set(reflect.Append(fieldVal, reflect.ValueOf(val)))
	return nil
}

func appendToBoolSlice(fieldVal reflect.Value, sliceVal string) error {
	var val bool
	switch sliceVal {
	case "true":
		val = true
	case "false":
		val = false
	default:
		return boolErr
	}
	fieldVal.Set(reflect.Append(fieldVal, reflect.ValueOf(val)))
	return nil
}

func unmarshalSlice(fieldVal reflect.Value, structField reflect.StructField, env environ, failOnUndefined bool) error {
	envKey := getEnvKey(structField)
	envNames := make([]string, 0)

	for envName, _ := range env {
		if strings.HasPrefix(envName, envKey) {
			envNames = append(envNames, envName)
		}
	}
	sort.Strings(envNames)
	if failOnUndefined && len(envNames) == 0 {
		return undefinedInEnvironError(envKey)
	}

	var err error
	for _, envName := range envNames {
		val, ok := env[envName]
		if !ok {
			continue
		}
		switch structField.Type.Elem().Kind() {
		case reflect.String:
			err = appendToStringSlice(fieldVal, val)
		case reflect.Int:
			err = appendToIntSlice(fieldVal, val)
		case reflect.Bool:
			err = appendToBoolSlice(fieldVal, val)
		}
		if err != nil {
			return err
		}

	}
	return nil
}

func unmarshalSingleField(fieldVal reflect.Value, structField reflect.StructField, env environ, failOnUndefined bool) error {
	if !fieldVal.CanSet() { // unexported field can not be set
		return nil
	}
	switch structField.Type.Kind() {
	case reflect.Int:
		return unmarshalInt(fieldVal, structField, env, failOnUndefined)
	case reflect.String:
		return unmarshalString(fieldVal, structField, env, failOnUndefined)
	case reflect.Bool:
		return unmarshalBool(fieldVal, structField, env, failOnUndefined)
	case reflect.Slice:
		return unmarshalSlice(fieldVal, structField, env, failOnUndefined)
	}
	return nil
}

func unmarshalAllStructFields(structVal reflect.Value, env environ, failOnUndefined bool) error {
	for i := 0; i < structVal.NumField(); i++ {
		if err := unmarshalSingleField(structVal.Field(i), structVal.Type().Field(i), env, failOnUndefined); err != nil {
			return err
		}
	}
	return nil
}

func getStructValue(v interface{}) reflect.Value {
	str := reflect.ValueOf(v)
	for {
		if str.Kind() == reflect.Struct {
			break
		}
		str = str.Elem()
	}
	return str
}

func makeSureValueIsInitialized(v interface{}) {
	if reflect.TypeOf(v).Elem().Kind() != reflect.Ptr {
		return
	}
	if reflect.ValueOf(v).Elem().IsNil() {
		reflect.ValueOf(v).Elem().Set(reflect.New(reflect.TypeOf(v).Elem().Elem()))
	}
}

func makeSureTypeIsSupported(v interface{}) (reflect.Type, error) {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return nil, errors.New("we need a pointer")
	}
	if reflect.TypeOf(v).Elem().Kind() == reflect.Ptr && reflect.TypeOf(v).Elem().Elem().Kind() == reflect.Struct {
		return reflect.TypeOf(v).Elem().Elem(), nil
	} else if reflect.TypeOf(v).Elem().Kind() == reflect.Struct && reflect.ValueOf(v).Elem().CanAddr() {
		return reflect.TypeOf(v).Elem(), nil
	}
	return nil, errors.New("we need a pointer to struct or pointer to pointer to struct")
}

func isSupportedStructField(k reflect.StructField) bool {
	switch k.Type.Kind() {
	case reflect.String:
		return true
	case reflect.Bool:
		return true
	case reflect.Int:
		return true
	case reflect.Slice:
		switch k.Type.Elem().Kind() {
		case reflect.String:
			return true
		case reflect.Bool:
			return true
		case reflect.Int:
			return true
		default:
			return false
		}

	default:
		return false
	}

}

func makeSureStructFieldTypesAreSupported(structType reflect.Type) error {
	for i := 0; i < structType.NumField(); i++ {
		if !isSupportedStructField(structType.Field(i)) {
			return fmt.Errorf("unsupported struct field type: %v", structType.Field(i).Type)
		}
	}
	return nil
}

type environ map[string]string

func getAllEnvironNames(envList []string) (map[string]struct{}, error) {
	envNames := make(map[string]struct{})

	for _, kv := range envList {
		split := strings.SplitN(kv, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("unknown environ condition - env variable not in k=v format: %v", kv)
		}
		envNames[split[0]] = struct{}{}
	}

	return envNames, nil
}

func newEnviron() (environ, error) {

	envNames, err := getAllEnvironNames(os.Environ())
	if err != nil {
		return nil, err
	}

	env := make(environ)

	for name, _ := range envNames {
		env[name] = os.ExpandEnv(os.Getenv(name))
	}

	return env, nil
}
