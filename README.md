# envcfg [![Build Status](https://travis-ci.org/tomazk/envcfg.svg?branch=master)](https://travis-ci.org/tomazk/envcfg)

Un-marshaling environment variables to Go structs

## Getting Started

Let's set a bunch of environment variables and then run your go app
```bash
#!/usr/bin/env bash
export DEBUG="false"
export DB_HOST="localhost"
export DB_PORT="8012"

./your_go_app 
```
Within your Go app do
```go
import "github.com/tomazk/envcfg"

// declare a type that will hold your env variables
type Cfg struct {
	DEBUG   bool
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
```
## Installation

```
$ go get github.com/tomazk/envcfg
```

## Motivation

As per **[12 factor app manifesto](http://12factor.net/)** configuration of an app should be stored in the [environment](http://12factor.net/config) since it varies between environments by nature. This convention is also dicated and popularized by emerging technologies like docker and cloud platforms. 

Instead of having a bunch of `os.Getenv("ENV_VAR")` buried deep in your code when configuring clients and services, **`envcfg`** encourages you to:

1. **define a struct type** that will hold your environment variables and serve as documentation which env variables must be configured
2. use **`envcfg.Unmarshal`** to read your env variables and unmarhsal them to an object that now holds your configuration of an app
3. use **`envcfg.ClearEnvVars`** to unset env variables, removing potential vulnerability of passing secrets to unsafe child processes or vendor libraries that assume you're not storing unsafe values in the environment

## Documentation

### `envcfg.Unmarshal`

`func Unmarshal(v interface{}) error` can recieve a reference to an object or even a reference to a pointer:

```go
var val2 StructType
envcfg.Unmarshal(&val2)

var val1 *StructType 
envcfg.Unmarshal(&val1) // val1 will be initialized
```

#### Supported Struct Field Types 

`envcfg.Unmarshal` supports `int`, `string`, `bool` and `[]int`, `[]string`, `[]bool` types of fields wihin a struct. In addition, fields that satisfy the `encoding.TextUnmarshaler` interface are also supported. `envcfg.Unmarshal` will return nil if a valid struct was passed or return an error if not.

```go
type StructType struct {
	INT           int
	BOOL          bool
	STRING        string
	SLICE_STRING  []string
	SLICE_BOOL    []bool
	SLICE_INT     []int
	CUSTOM_TYPE   MyType
}

type MyType struct{}
func (mt *MyType) UnmarshalText(text []byte) error {
	...
}
```
#### Validation
`envcfg.Unmarshal` also spares you from writing type validation code:

```go
type StructType struct {
	SHOULD_BE_INT int
}
```
If you'll pass `export SHOULD_BE_INT="some_string_value"` to your application `envcfg.Unmarshal` will return an error.

##### Undefined Variables
`EnvUnmarshaler.Unmarshal` can also be set to fail on undefined variables:

```go
package main

import (
	"github.com/tomazk/envcfg"
	"log"
)

type StructType struct {
	MUST_BE_IN_ENVIRONMENT int
}

func main() {
	u := envcfg.NewUnmarshaler().FailOnUndefined()
	err := u.Unmarshal(&StructType)
	if err {
		log.Fatal(err)
	}
}
```
If you run this program without `MUST_BE_IN_ENVIRONMENT` defined in the environment, then it will exit with the error

#### Struct Tags for Custom Mapping of env Variables
You can also use struct field tags to map env variables to fields wihin a struct
```bash
export MY_ENV_VAR=1
```
```go
type StructType struct {
	Field int `envcfg:"MY_ENV_VAR"`
}
```
#### Slices Support
`envcfg.Unmarshal` also supports `[]int`, `[]string`, `[]bool` slices. Values of the slice are ordered in respect to env name suffix. See example below.
```bash
export CASSANDRA_HOST_1="192.168.0.20" # *_1 will come as the first element of the slice
export CASSANDRA_HOST_2="192.168.0.21"
export CASSANDRA_HOST_3="192.168.0.22"
```
```go
type StructType struct {
	CASSANDRA_HOST []string
}
func main() {
	var config StructType
	envcfg.Unmarshal(&config)
	// config.CASSANDRA_HOST is now set to []string{"192.168.0.20", "192.168.0.21", "192.168.0.22"} 
}
```
### `envcfg.ClearEnvVars`

`func ClearEnvVars(v interface{}) error` recieves a reference to the same struct you've passed to `envcfg.Unmarshal` and it will unset any environment variables listed in the struct. Except for those that you want to keep and are tagged with `envcfgkeep:""` struct field tag. It will throw an error on unsupported types.

```bash
export SECRET_AWS_KEY="foobar" 
export PORT="8080" 
```
```go
type StructType struct {
	SECRET_AWS_KEY string
	PORT           int    `envcfgkeep:""`
}
func main() {
	var config StructType
	envcfg.ClearEnvVars(&config)
	// it will unset SECRET_AWS_KEY but keep env variable PORT
}
```


## Contributing
Send me a pull request and make sure tests pass on [travis](https://travis-ci.org/tomazk/envcfg/).

## Tests

Package comes with an extensive test suite that's continuously run on travis against go versions: 1.3, 1.4, 1.5, 1.6 and the development tip.
```
$ go test github.com/tomazk/envcfg
```

## Licence

See LICENCE file
