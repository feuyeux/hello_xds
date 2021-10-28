#!/bin/bash
cd "$(
  cd "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)/" || exit
go run xds/xds_server.go --upstream_port=50051 --upstream_port=50052