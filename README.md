# Hello gRPC xDS

> base from https://github.com/salrashid123/grpc_xds

## build

```bash
go mod tidy
```

## run p2p without xds

### app server1

```bash
go run app/grpc_server.go --grpcport :50051 --servername server1
```

### app client_dns

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
go run app/grpc_client_dns.go
```

## run lb with xds

### app server1

```bash
go run app/grpc_server.go --grpcport :50051 --servername server1
```

### app server2

```bash
go run app/grpc_server.go --grpcport :50052 --servername server2
```

### xds server

```bash
go run xds/xds_server.go --upstream_port=50051 --upstream_port=50052
```

### app client_xds

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
export GRPC_XDS_BOOTSTRAP=`pwd`/xds_bootstrap.json
go run app/grpc_client_xds.go
```

### debug xds client
```bash
go install -v github.com/grpc-ecosystem/grpcdebug@latest
grpcdebug localhost:50053 xds status
```