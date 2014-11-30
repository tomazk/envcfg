package envcfg

import(
    "fmt"
    "strings"
    "os"
    "reflect"
    "errors"
)


func Unmarshal(v interface{}) error {
    structType, err := makeSureTypeIsSupported(v)
    if err != nil {
        return err
    }
    if err := makeSureStructFieldTypesAreSupported(structType); err != nil {
        return err
    }
    makeSureValueIsInitialized(v)
    strct := getStructValue(v)


   



    return nil 
}



func getStructValue(v interface{}) reflect.Value {
    str := reflect.ValueOf(v) 
    for {
        fmt.Println(str.Kind())
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



