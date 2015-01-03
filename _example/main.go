package main

import (
	"fmt"
	"github.com/tomazk/envcfg"
	"os"
)

type Cfg struct {
	DEBUG bool

	CASSANDRA_PORT  int
	CASSANDRA_HOSTS []string

	STATSD_HOST string
	STATSD_PORT int
}

func main() {
	var config Cfg
	if err := envcfg.Unmarshal(&config); err != nil {
		fmt.Println("error when Unmarshal")
		os.Exit(1)
	}
	fmt.Printf("%#v\n", config)
}
