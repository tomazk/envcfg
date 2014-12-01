# envcfg [![wercker status](https://app.wercker.com/status/5ecc1b6d732792c4112e05d2d69334f3/s/master "wercker status")](https://app.wercker.com/project/bykey/5ecc1b6d732792c4112e05d2d69334f3)


Un-marshaling environment variables to Go structs

## Getting Started

Let's set a bunch of environment variables and then run your go app
```bash
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
var val1 *StructType
envcfg.Unmarshal(&val1)

var val2 StructType
envcfg.Unmarshal(&val2)
```






