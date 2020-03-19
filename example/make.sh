#!/bin/sh

go build github.com/ntons/tbl-go/tblmaker

$(dirname $0)/tblmaker -p example -I proto -i "common.proto" -P proto -O tbl xlsx

