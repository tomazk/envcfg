# envcfg [![wercker status](https://app.wercker.com/status/5ecc1b6d732792c4112e05d2d69334f3/s/master "wercker status")](https://app.wercker.com/project/bykey/5ecc1b6d732792c4112e05d2d69334f3) [![Build Status](https://travis-ci.org/tomazk/envcfg.svg?branch=master)](https://travis-ci.org/tomazk/envcfg)

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
	DEBUG bool
	DB_PORT int
	DB_HOST string
}

func main() {
	var config Cfg
	envcfg.Unmarshal(&config)
	// config is now set to Config{DEBUG: false, DB_PORT: 8012, DB_HOST: "localhost"}
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

`envcfg.Unmarshal` supports `int`, `string`, `bool` and `[]int`, `[]string`, `[]bool` types of fields wihin a struct. It will return nil if a valid struct was passed or return an error if not.
```go
type StructType struct {
	INT           int
	BOOL          bool
	STRING        string
	SLICE_STRING  []string
	SLICE_BOOL    []bool
	SLICE_INT     []int
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
## Contributing
Send me a pull request and make sure tests pass on [travis](https://travis-ci.org/tomazk/envcfg/).

## Tests

Package comes with an extensive test suite that's continously run on travis against go versions: 1.3, 1.4 and the development tip.
```
$ go test github.com/tomazk/envcfg
```

## Licence

See LICENCE file
