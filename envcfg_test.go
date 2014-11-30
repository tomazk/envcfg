package envcfg

import (
    "testing"
    "os"
    "reflect"
    _ "fmt"
)

func init() {
    os.Clearenv()
}

func setEnv(t *testing.T, k string, v string) {
    err := os.Setenv(k, v)
    if err != nil {
        t.Fatal("error when setting env")
    }
}


func TestEnvironNames(t *testing.T) {
    
    names, err := getAllEnvironNames([]string{"key1=val", "key2=", "key1=anotherval"})
    if err != nil {
        t.FailNow()
    }
    want := map[string]struct{}{"key1":struct{}{}, "key2":struct{}{}} 
    if !reflect.DeepEqual(names, want) {
        t.FailNow()
    }

    names, err = getAllEnvironNames([]string{"key1=val", "key2"})
    if err == nil && names != nil {
        t.FailNow()
    }
}

func TestNewEnviron(t *testing.T) {
    setEnv(t, "key1", "val")
    setEnv(t, "key2", "")
    setEnv(t, "key3", "")
    setEnv(t, "key3", "val3")
    defer os.Clearenv()

    env, err := newEnviron()
    if err != nil {
        t.Fatal("error when calling newEnviron")
    }
    want := environ{"key1": "val", "key2": "", "key3": "val3"}
    if !reflect.DeepEqual(want, env) {
        t.Fatalf("env not eq to want %#v compared to have %#v", want, env)
    }
}

type cfgValid1 struct {
    STRING string
    INT int
    BOOL bool
    STRING_SLICE []string
    INT_SLICE []int
    BOOL_SLICE []bool
}

type validType int
type cfgValid2 struct {
    INT_SLICE []validType
    INT validType
}

type cfgInvalid1 struct {
    FLOAT float64
}

type cfgInvalid2 struct {
    FLOAT_SLICE []float64
}



func TestUnmarshalValidateType(t *testing.T) {
    
    var i int
    if err := Unmarshal(i); err == nil {
        t.Fatal("should fail if we don't pass addressable value")
    }

    var p *cfgValid1
    if err := Unmarshal(p); err == nil {
        t.Fatal("pointer type: should fail if we don't pass addressable value")
    }
    if err := Unmarshal(&p); err != nil {
        t.Fatal("pointer type: should not fail since passed an addressable value")
    }

    var v cfgValid1
    if err := Unmarshal(&v); err != nil {
        t.Fatal("should not fail since we passed a valid value addressable")
    }
    if err := Unmarshal(v); err == nil {
        t.Fatal("should fail since we did not pass an addressable value")
    }

    var v1 cfgValid2
    if err := Unmarshal(&v1); err != nil {
        t.Fatal("should not fail since we passed another valid value")
    }

    var inv1 cfgInvalid1
    if err := Unmarshal(&inv1); err == nil {
        t.Log(err)
        t.Fatal("should fail due to invalid struct type")
    }

    var inv2 cfgInvalid2
    if err := Unmarshal(&inv2); err == nil {
        t.Log(err)
        t.Fatal("should fail due to invalid struct type - second case")
    }
}

func TestUnmarshalInit(t *testing.T) {
    var p *cfgValid1
    Unmarshal(&p)
    if !reflect.DeepEqual(p, new(cfgValid1)) {
        t.Fatal("should be initialized")
    }

    var v cfgValid1
    Unmarshal(&v) // shouldn't panic
}

func TestUnmarshalSetFields(t *testing.T) {
    var p *cfgValid1
    Unmarshal(&p)
}




