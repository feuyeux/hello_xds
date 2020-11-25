docker run --network=host grpc_xds_hello:1.0.0 \
./grpc_server --grpcport :50051 --servername server1

docker run --network=host grpc_xds_hello:1.0.0 \
./grpc_server --grpcport :50052 --servername server2

docker run --network=host grpc_xds_hello:1.0.0 \
./grpc_client --host dns:///hello-server:50051

docker run --network=host grpc_xds_server:1.0.0 \
./main --upstream_port=50051 --upstream_port=50052

docker run --network=host -e GRPC_XDS_BOOTSTRAP=/xds_bootstrap.json \
grpc_xds_hello:1.0.0 \
./grpc_client --host xds:///hello-service