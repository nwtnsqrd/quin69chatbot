#!/usr/bin/env bash

set -xe

MODE=${1:-"offlinechat"}

go build -v -o quinbot
./quinbot --mode $MODE