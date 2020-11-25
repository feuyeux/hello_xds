#!/usr/bin/env bash
SCRIPT_PATH="$(
  cd "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)/"
cd "$SCRIPT_PATH" || exit
cp -R ../app/ app
docker build -f hello.dockerfile -t grpc_xds_hello:1.0.0 .
rm -rf app
cp -R ../xds/ xds
docker build -f xds.dockerfile -t grpc_xds_server:1.0.0 .
rm -rf xds