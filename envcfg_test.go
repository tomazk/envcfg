package envcfg

import (
	_ "fmt"
	"os"
	"reflect"
	"testing"
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
	want := map[string]struct{}{"key1": struct{}{}, "key2": struct{}{}}
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
	STRING       string
	INT          int
	BOOL         bool
	STRING_SLICE []string
	INT_SLICE    []int
	BOOL_SLICE   []bool
}

type validType int
type cfgValid2 struct {
	INT_SLICE []validType
	INT       validType
}

type cfgInvalid1 struct {
	FLOAT float64
}

type cfgInvalid2 struct {
	FLOAT_SLICE []float64
}

func TestUnmarshalValidateType(t *testing.T) {

	var i int
	if err := Unmarshal(i, false); err == nil {
		t.Fatal("should fail if we don't pass addressable value")
	}

	var p *cfgValid1
	if err := Unmarshal(p, false); err == nil {
		t.Fatal("pointer type: should fail if we don't pass addressable value")
	}
	if err := Unmarshal(&p, false); err != nil {
		t.Fatal("pointer type: should not fail since passed an addressable value")
	}

	var v cfgValid1
	if err := Unmarshal(&v, false); err != nil {
		t.Fatal("should not fail since we passed a valid value addressable")
	}
	if err := Unmarshal(v, false); err == nil {
		t.Fatal("should fail since we did not pass an addressable value")
	}

	var v1 cfgValid2
	if err := Unmarshal(&v1, false); err != nil {
		t.Fatal("should not fail since we passed another valid value")
	}

	var inv1 cfgInvalid1
	if err := Unmarshal(&inv1, false); err == nil {
		t.Log(err)
		t.Fatal("should fail due to invalid struct type")
	}

	var inv2 cfgInvalid2
	if err := Unmarshal(&inv2, false); err == nil {
		t.Log(err)
		t.Fatal("should fail due to invalid struct type - second case")
	}
}

func TestClearEnvVarsSupportedType(t *testing.T) {
	var p cfgValid1
	if err := ClearEnvVars(p); err == nil {
		t.Fatal("should fail since we didn't pass a reference")
	}

	var p1 cfgInvalid1
	if err := ClearEnvVars(&p1); err == nil {
		t.Fatal("should fail since we passed in an invalid struct")
	}
}

type ClearEnvVarsType1 struct {
	THIS string
	THAT string `envcfgkeep:""`
	Foo  string `envcfg:"FOO"`
	Bar  string `envcfgkeep:"" envcfg:"BAR"`
}

func TestClearEnvVars(t *testing.T) {
	setEnv(t, "THIS", "1")
	setEnv(t, "THAT", "1")
	setEnv(t, "FOO", "1")
	setEnv(t, "BAR", "1")
	defer os.Clearenv()

	env, _ := newEnviron()
	if !reflect.DeepEqual(env, environ{
		"FOO":  "1",
		"BAR":  "1",
		"THIS": "1",
		"THAT": "1",
	}) {
		t.FailNow()
	}

	ClearEnvVars(&ClearEnvVarsType1{})

	env, _ = newEnviron()
	if !reflect.DeepEqual(env, environ{
		"FOO":  "",
		"BAR":  "1",
		"THIS": "",
		"THAT": "1",
	}) {
		t.FailNow()
	}

}

func TestUnmarshalInit(t *testing.T) {
	var p *cfgValid1
	Unmarshal(&p, false)
	if !reflect.DeepEqual(p, new(cfgValid1)) {
		t.Fatal("should be initialized")
	}

	var v cfgValid1
	Unmarshal(&v, false) // shouldn't panic
}

type IntType struct {
	INT   int
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
	Unmarshal(&i, false)
	if !reflect.DeepEqual(IntType{0, 1, 2, 3, 3}, i) {
		t.Fatal("should be eq")
	}

	var p *IntType
	Unmarshal(&p, false)
	if !reflect.DeepEqual(&IntType{0, 1, 2, 3, 3}, p) {
		t.Fatal("should be eq")
	}

	setEnv(t, "INT_1", "invalid")
	if err := Unmarshal(&i, false); err == nil {
		t.Fatal("should throw an error since we passed an invalid int value")
	}
}

func TestUnmarshalIntFailOnUndefined(t *testing.T) {
	setEnv(t, "INT_1", "1")
	setEnv(t, "INT_2", "2")
	defer os.Clearenv()

	var i IntType
	if err := Unmarshal(&i, true); err == nil {
		t.Fatal("undefined variables should cause error")
	}
}

type StringType struct {
	STR   string
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
	Unmarshal(&s, false)
	if !reflect.DeepEqual(StringType{"", "s1", "s2", "s3", "s3"}, s) {
		t.Fatal("should be equal")
	}
}

func TestUnmarshalStringFailOnUndefined(t *testing.T) {
	setEnv(t, "STR_1", "s1")
	setEnv(t, "STR_2", "s2")
	setEnv(t, "LABEL_STR", "s3")
	defer os.Clearenv()

	var s StringType
	if err := Unmarshal(&s, true); err == nil {
		t.Fatal("undefined variables should cause error")
	}
}

type BoolType struct {
	BOOL   bool
	BOOL_1 bool
	BOOL_2 bool `envcfg:"LABEL_BOOL"`
}

func TestUnmarshalBool(t *testing.T) {
	setEnv(t, "BOOL_1", "true")
	setEnv(t, "LABEL_BOOL", "true")
	defer os.Clearenv()

	var b BoolType
	Unmarshal(&b, false)
	if !reflect.DeepEqual(BoolType{false, true, true}, b) {
		t.Log(b)
		t.Fatal("should be equal")
	}

	setEnv(t, "LABEL_BOOL", "invalid")
	if err := Unmarshal(&b, false); err == nil {
		t.Fatal("should fail")
	}
}

func TestUnmarshalBoolFailOnUndefined(t *testing.T) {
	setEnv(t, "BOOL_1", "true")
	setEnv(t, "LABEL_BOOL", "true")
	defer os.Clearenv()

	var b BoolType
	if err := Unmarshal(&b, true); err == nil {
		t.Fatal("undefined variables should cause error")
	}
}

type SliceType struct {
	SLICE_STR  []string
	SLICE_INT  []int
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
	Unmarshal(&s, false)
	if !reflect.DeepEqual(s, SliceType{[]string{"foo", "bar"}, []int{1, 2}, []bool{true, false}}) {
		t.Fatal("should be equal")
	}
}

func TestUnmarshalSliceFail(t *testing.T) {
	defer os.Clearenv()
	setEnv(t, "SLICE_BOOL_1", "true")
	setEnv(t, "SLICE_BOOL_2", "invalid")
	var s SliceType

	if err := Unmarshal(&s, false); err == nil {
		t.Fatal("should fail on an invalid bool value")
	}
	os.Clearenv()

	setEnv(t, "SLICE_INT_1", "invalid")
	if err := Unmarshal(&s, false); err == nil {
		t.Fatal("shoud fail on invalid int")
	}
}

func TestUnmarshalSliceFailUndefined(t *testing.T) {
	defer os.Clearenv()
	setEnv(t, "SLICE_STR_1", "foo")
	setEnv(t, "SLICE_STR_2", "bar")
	setEnv(t, "SLICE_INT_1", "1")
	setEnv(t, "SLICE_INT_2", "2")
	var s SliceType

	if err := Unmarshal(&s, true); err == nil {
		t.Fatal("should fail on missing environment variable")
	}
}

type GeneralTest struct {
	some_int int // unexported - will not be set
	SOME_INT int

	SOME_STR string

	SOME_BOOL bool

	SOME_SLICE_BOOL   []bool
	SOME_SLICE_INT    []int
	SOME_SLICE_STRING []string

	SOME_UNSET_FIELD string
}

func TestGeneral(t *testing.T) {
	defer os.Clearenv()
	setEnv(t, "some_int", "1")
	setEnv(t, "SOME_INT", "1")
	setEnv(t, "SOME_STR", "1")
	setEnv(t, "SOME_BOOL", "true")
	setEnv(t, "SOME_SLICE_BOOL_a", "true")
	setEnv(t, "SOME_SLICE_INT", "1")
	setEnv(t, "SOME_SLICE_INT_1", "1")
	setEnv(t, "SOME_SLICE_INT_2", "5")
	setEnv(t, "BAR", "bar")
	setEnv(t, "SOME_SLICE_STRING", "foo${BAR}")

	var gt GeneralTest
	if err := Unmarshal(&gt, false); err != nil {
		t.Fatal("should not fail")
	}
	want := GeneralTest{
		SOME_INT:          1,
		SOME_STR:          "1",
		SOME_BOOL:         true,
		SOME_SLICE_BOOL:   []bool{true},
		SOME_SLICE_INT:    []int{1, 1, 5},
		SOME_SLICE_STRING: []string{"foobar"},
	}
	if !reflect.DeepEqual(gt, want) {
		t.Fatal("should be eq")
	}

}
