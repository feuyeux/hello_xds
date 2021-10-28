#!/bin/bash
cd "$(
  cd "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)/" || exit
go run app/grpc_server.go --grpcport :50051 --servername server1
