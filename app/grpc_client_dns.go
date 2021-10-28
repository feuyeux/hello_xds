package main

import (
	_ "google.golang.org/grpc/resolver" // use for "dns:///be.cluster.local:50051"
	"hello_xds/app/client"
)

func main() {
	address := "dns:///be.cluster.local:50051"
	client.Run(&address)
}
