package envcfg

import(
    "fmt"
    "strings"
    "os"
    "reflect"
    "errors"
    "strconv"
    "sort"
)

const structTag = "envcfg" 


func Unmarshal(v interface{}) error {
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

    if err := unmarshalAllStructFields(structVal, env); err != nil {
        return err
    }

    return nil 
}

func getEnvKey(structField reflect.StructField) string {
    if tag := structField.Tag.Get(structTag); tag != "" {
        return tag
    }
    return structField.Name
}


func unmarshalInt(fieldVal reflect.Value, structField reflect.StructField, env environ) error {
    val, ok :=  env[getEnvKey(structField)]
    if !ok {
        return nil
    }

    i, err := strconv.Atoi(val)
    if err != nil {
        return err
    }

    fieldVal.SetInt(int64(i))
    return nil 
}

var boolErr error = errors.New("pass string 'true' or 'false' for boolean fields")

func unmarshalBool(fieldVal reflect.Value, structField reflect.StructField, env environ) error {
    val, ok :=  env[getEnvKey(structField)]
    if !ok {
        return nil
    }

    var vbool bool
    switch val {
        case "true": vbool = true
        case "false": vbool = false
        default: return boolErr   
    }

    fieldVal.SetBool(vbool)
    return nil
}

func unmarshalString(fieldVal reflect.Value, structField reflect.StructField, env environ) error {
    val, ok :=  env[getEnvKey(structField)]
    if !ok {
        return nil
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
        case "true": val = true
        case "false": val = false
        default: return boolErr
    }
    fieldVal.Set(reflect.Append(fieldVal, reflect.ValueOf(val)))
    return nil
}

func unmarshalSlice(fieldVal reflect.Value, structField reflect.StructField, env environ) error {
    envKey := getEnvKey(structField)
    envNames := make([]string, 0)

    for envName, _ := range env {
        if strings.HasPrefix(envName, envKey) {
            envNames = append(envNames, envName)
        }
    }
    sort.Strings(envNames)

    var err error
    for _ , envName := range envNames {
        val, ok := env[envName]
        if !ok {
            continue
        }
        switch structField.Type.Elem().Kind() {
            case reflect.String: err = appendToStringSlice(fieldVal, val)
            case reflect.Int: err = appendToIntSlice(fieldVal, val)
            case reflect.Bool: err = appendToBoolSlice(fieldVal, val)
        }
        if err != nil {
            return err
        }

    }
    return nil
}

func unmarshalSingleField(fieldVal reflect.Value, structField reflect.StructField, env environ) error {
    switch structField.Type.Kind() {
        case reflect.Int: return unmarshalInt(fieldVal, structField, env)
        case reflect.String: return unmarshalString(fieldVal, structField, env)
        case reflect.Bool: return unmarshalBool(fieldVal, structField, env)
        case reflect.Slice: return unmarshalSlice(fieldVal, structField, env)
    }
    return nil 
}

func unmarshalAllStructFields(structVal reflect.Value, env environ) error {
    for i := 0; i < structVal.NumField(); i++ {
        if err := unmarshalSingleField(structVal.Field(i), structVal.Type().Field(i), env); err != nil {
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
        case reflect.String: return true
        case reflect.Bool: return true
        case reflect.Int: return true
        case reflect.Slice: 
            switch k.Type.Elem().Kind() {
                case reflect.String: return true
                case reflect.Bool: return true
                case reflect.Int: return true
                default: return false
            }

        default: return false
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
