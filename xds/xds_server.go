package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"hello_xds/xds/proxy"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	ep "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	lv2 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	Ads             = "ads"
	backendHostName = "be.cluster.local"
	//must match what is specified at xds:///
	listenerName    = "be-srv"
	routeConfigName = "be-srv-route"
	clusterName     = "be-srv-cluster"
	virtualHostName = "be-srv-vs"
)

// UpstreamPorts is a type that implements flag.Value interface
type UpstreamPorts []int

// String is a method that implements the flag.Value interface
func (u *UpstreamPorts) String() string {
	// See: https://stackoverflow.com/a/37533144/609290
	return strings.Join(strings.Fields(fmt.Sprint(*u)), ",")
}

// Set is a method that implements the flag.Value interface
func (u *UpstreamPorts) Set(port string) error {
	log.Printf("[UpstreamPorts] %s", port)
	i, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	*u = append(*u, i)
	return nil
}

var (
	debug         bool
	port          uint
	mode          string
	version       int32
	config        cachev3.SnapshotCache
	upstreamPorts UpstreamPorts
)

var nodeId string

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.StringVar(&mode, "ads", Ads, "Management server type (ads, xds, rest)")
	// Converts repeated flags (e.g. `--upstream_port=50051 --upstream_port=50052`) into a []int
	flag.Var(&upstreamPorts, "upstream_port", "list of upstream gRPC servers")
}

/**/
const grpcMaxConcurrentStreams = 1000

// RunManagementServer starts an xDS server at the given port.
func RunManagementServer(ctx context.Context, server xds.Server, port uint) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	// register services
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

	log.WithFields(log.Fields{"port": port}).Info("management server listening")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Error(err)
		}
	}()
	<-ctx.Done()

	grpcServer.GracefulStop()
}

func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	ctx := context.Background()
	log.Printf("Starting Control Plane")

	signal := make(chan struct{})
	cb := &proxy.Callbacks{
		Signal:   signal,
		Fetches:  0,
		Requests: 0,
	}
	config = cachev3.NewSnapshotCache(true, cachev3.IDHash{}, nil)
	srv := xds.NewServer(ctx, config, cb)
	go RunManagementServer(ctx, srv, port)
	<-signal

	cb.Report()
	nodeId = config.GetStatusKeys()[0]
	log.Infof("Creating NodeID %s", nodeId)
	if upstreamPorts == nil {
		upstreamPorts.Set("50051")
		upstreamPorts.Set("50052")
	}
	var wg sync.WaitGroup
	trys := 10
	wg.Add(trys)
	for i := 0; i < trys; i++ {
		for _, v := range upstreamPorts {
			// ENDPOINT
			endpoints, isUp := buildEDS(v)
			if isUp {
				atomic.AddInt32(&version, 1)
				log.Infof("Creating snapshot Version " + fmt.Sprint(version))
				// CLUSTER
				clusters := buildCDS()
				// RDS
				routes := buildRDS()
				// LISTENER
				listeners := buildLDS()
				var runtimes []types.Resource
				var secrets []types.Resource
				snap := cachev3.NewSnapshot(fmt.Sprint(version), endpoints, clusters, routes, listeners, runtimes, secrets)
				config.SetSnapshot(nodeId, snap)
			}
			time.Sleep(10 * time.Second)
		}
		wg.Done()
	}
	wg.Wait()
}

/**/
func buildLDS() []types.Resource {
	log.Infof("Creating LISTENER " + listenerName)
	hcRds := &hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			RouteConfigName: routeConfigName,
			ConfigSource: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
	}
	manager := &hcm.HttpConnectionManager{
		CodecType:      hcm.HttpConnectionManager_AUTO,
		RouteSpecifier: hcRds,
	}
	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	l := []types.Resource{
		&listener.Listener{
			Name: listenerName,
			ApiListener: &lv2.ApiListener{
				ApiListener: pbst,
			},
		}}
	return l
}

func buildRDS() []types.Resource {
	log.Infof("Creating RDS " + virtualHostName)
	vh := &route.VirtualHost{
		Name:    virtualHostName,
		Domains: []string{listenerName},
		Routes: []*route.Route{{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "",
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}}}

	rds := []types.Resource{
		&route.RouteConfiguration{
			Name:         routeConfigName,
			VirtualHosts: []*route.VirtualHost{vh},
		},
	}
	return rds
}

func buildCDS() []types.Resource {
	log.Infof("Creating CLUSTER " + clusterName)
	cls := []types.Resource{
		&cluster.Cluster{
			Name:                 clusterName,
			LbPolicy:             cluster.Cluster_ROUND_ROBIN,
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
			EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
				EdsConfig: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{},
				},
			},
		},
	}
	return cls
}

func buildEDS(v int) ([]types.Resource, bool) {
	isUp := true
	log.Infof("Creating ENDPOINT for remoteHost:port %s:%d", backendHostName, v)
	hst := &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  backendHostName,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(v),
			},
		},
	}}

	identifier := &ep.LbEndpoint_Endpoint{
		Endpoint: &ep.Endpoint{
			Address: hst,
		}}

	// read from snapshot
	snapshot, err := config.GetSnapshot(nodeId)
	if err == nil {
		resources := snapshot.GetResources(resource.EndpointType)
		if resources != nil {
			// get eds config
			res := resources[clusterName]
			assignment := res.(*endpoint.ClusterLoadAssignment)
			endpoints := assignment.GetEndpoints()
			lbEndpoints := endpoints[0].GetLbEndpoints()
			currentPortValue := uint32(v)
			for _, lbEndpoint := range lbEndpoints {
				portValue := lbEndpoint.GetEndpoint().GetAddress().GetSocketAddress().GetPortValue()
				if portValue == currentPortValue {
					isUp = false
					break
				}
			}
			if isUp {
				lbEp := &ep.LbEndpoint{
					HostIdentifier: identifier,
					HealthStatus:   core.HealthStatus_HEALTHY,
				}
				endpoints[0].LbEndpoints = append(lbEndpoints, lbEp)
			}
			log.Infof("EPS:%+v", endpoints)
			return []types.Resource{res}, isUp
		}
	}

	//eds := []cache.Resource{
	eds := []types.Resource{
		&endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*ep.LocalityLbEndpoints{{
				Locality: &core.Locality{
					Region: "us-central1",
					Zone:   "us-central1-a",
				},
				Priority:            0,
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: uint32(1000)},
				LbEndpoints: []*ep.LbEndpoint{
					{
						HostIdentifier: identifier,
						HealthStatus:   core.HealthStatus_HEALTHY,
					},
				},
			}},
		},
	}
	return eds, isUp
}
