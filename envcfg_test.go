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


type IntType struct {
    INT int
    INT_1 int 
    INT_2 int `envcfg:""`
    INT_3 int `envcfg:"LABEL_INT"`
    INT_4 int `envcfg:"LABEL_INT"`
}

func TestUnmarshalSetInt(t *testing.T) {
    setEnv(t, "INT_1", "1")
    setEnv(t, "INT_2", "2")
    setEnv(t, "LABEL_INT", "3")
    defer os.Clearenv()

    var i IntType
    Unmarshal(&i)
    if !reflect.DeepEqual(IntType{0, 1, 2, 3, 3}, i) {
        t.Fatal("should be eq")
    }

    var p *IntType
    Unmarshal(&p)
    if !reflect.DeepEqual(&IntType{0, 1, 2, 3, 3}, p) {
        t.Fatal("should be eq")
    }

    setEnv(t, "INT_1", "invalid")
    if err := Unmarshal(&i); err == nil {
        t.Fatal("should throw an error since we passed an invalid int value")
    }
}


type StringType struct {
    STR string
    STR_1 string
    STR_2 string `envcfg:""`
    STR_3 string `envcfg:"LABEL_STR"`
    STR_4 string `envcfg:"LABEL_STR"`
}

func TestUnmarshalString(t *testing.T) {
    setEnv(t, "STR_1", "s1")
    setEnv(t, "STR_2", "s2")
    setEnv(t, "LABEL_STR", "s3")
    defer os.Clearenv()

    var s StringType
    Unmarshal(&s)
    if !reflect.DeepEqual(StringType{"", "s1","s2","s3", "s3"},s) {
        t.Fatal("should be equal")
    }
}


type BoolType struct {
    BOOL bool
    BOOL_1 bool
    BOOL_2 bool `envcfg:"LABEL_BOOL"`
}

func TestUnmarshalBool(t *testing.T) {
    setEnv(t, "BOOL_1", "true")
    setEnv(t, "LABEL_BOOL", "true")
    defer os.Clearenv()

    var b BoolType
    Unmarshal(&b)
    if !reflect.DeepEqual(BoolType{false, true, true}, b) {
        t.Log(b)
        t.Fatal("should be equal")
    }

    setEnv(t, "LABEL_BOOL", "invalid")
    if err := Unmarshal(&b); err == nil {
        t.Fatal("should fail")
    }
}

type SliceType struct {
    SLICE_STR []string
    SLICE_INT []int
    SLICE_BOOL []bool
}

func TestUnmarshalSlice(t *testing.T) {
    setEnv(t, "SLICE_STR_1", "foo")
    setEnv(t, "SLICE_STR_2", "bar")
    setEnv(t, "SLICE_INT_1", "1")
    setEnv(t, "SLICE_INT_2", "2")
    setEnv(t, "SLICE_BOOL_1", "true")
    setEnv(t, "SLICE_BOOL_2", "false")
    defer os.Clearenv()

    var s SliceType
    Unmarshal(&s)
    if !reflect.DeepEqual(s, SliceType{[]string{"foo", "bar"}, []int{1, 2}, []bool{true, false}}) {
        t.Fatal("should be equal")
    }
}


func TestUnmarshalSliceFail(t *testing.T) {
    defer os.Clearenv()
    setEnv(t, "SLICE_BOOL_1", "true")
    setEnv(t, "SLICE_BOOL_2", "invalid")
    var s SliceType

    if err := Unmarshal(&s); err == nil {
        t.Fatal("should fail on an invalid bool value")
    }
    os.Clearenv()

    setEnv(t, "SLICE_INT_1", "invalid")
    if err := Unmarshal(&s); err == nil {
        t.Fatal("shoud fail on invalid int")
    }
}
