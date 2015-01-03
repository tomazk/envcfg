#!/usr/bin/env bash

export DEBUG=true
export CASSANDRA_PORT=123
export CASSANDRA_HOSTS_1="192.168.0.20"
export CASSANDRA_HOSTS_2="192.168.0.21"
export LOCAL="local"
export STATSD_HOST="${LOCAL}host"
export STATSD_PORT=123

go run ./main.go