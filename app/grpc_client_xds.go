package main

import (
	//_ "google.golang.org/grpc/resolver" // use for "dns:///be.cluster.local:50051"
	_ "google.golang.org/grpc/xds" // use for xds-experimental:///be-srv
	"hello_xds/app/client"
)

func main() {
	address := "xds:///be-srv"
	client.Run(&address)
}
