#!/bin/bash
cd "$(
  cd "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)/" || exit
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
export GRPC_XDS_BOOTSTRAP=$(pwd)/xds_bootstrap.json
go run app/grpc_client_xds.go
