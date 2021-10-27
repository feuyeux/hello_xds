# Hello gRPC xDS
> base from https://github.com/salrashid123/grpc_xds

## build
```bash
go mod tidy
```

## run without xds
### app server1
```bash
go run app/grpc_server.go --grpcport :50051 --servername server1
```

### app client_dns
```bash
go run app/grpc_client_dns.go --host dns:///be.cluster.local:50051
```

## run with xds

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
export GRPC_XDS_BOOTSTRAP=`pwd`/xds_bootstrap.json
go run app/grpc_client_xds.go --host xds:///be-srv
```