package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"echo"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	grpcport   = flag.String("grpcport", "", "grpcport")
	servername = flag.String("servername", "server1", "grpcport")
	hs         *health.Server
)

/**/
type server struct {
}

func isGrpcRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc")
}

func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {
	log.Println("-> ", in.Name)
	return &echo.EchoReply{Message: *servername + " reply " + in.Name}, nil
}

func (s *server) SayHelloStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloStreamServer) error {
	log.Println("-->  ")
	stream.Send(&echo.EchoReply{Message: "Hello " + in.Name})
	stream.Send(&echo.EchoReply{Message: "Hello " + in.Name})
	return nil
}

/**/
type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Check request: %v", in)
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func main() {
	flag.Parse()
	if *grpcport == "" {
		fmt.Fprintln(os.Stderr, "missing -grpcport flag (:50051)")
		flag.Usage()
		os.Exit(2)
	}
	lis, err := net.Listen("tcp", *grpcport)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}
	s := grpc.NewServer(sopts...)
	echo.RegisterEchoServerServer(s, &server{})

	healthpb.RegisterHealthServer(s, &healthServer{})
	log.Println("Starting grpcServer")
	s.Serve(lis)
}
