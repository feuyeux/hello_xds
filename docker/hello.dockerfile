FROM golang:1.15 as build

RUN apt-get update -y && apt-get install -y build-essential wget unzip curl

ENV GO111MODULE=on
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

WORKDIR /app

ADD app /app
COPY protoc-3.2.0-linux-x86_64.zip /tmp/

RUN unzip /tmp/protoc-3.2.0-linux-x86_64.zip -d protoc3 && \
    mv protoc3/bin/* /usr/local/bin/ && \
    mv protoc3/include/* /usr/local/include/
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go mod download

RUN /usr/local/bin/protoc -I src/ --include_imports --include_source_info --descriptor_set_out=src/echo/echo.proto.pb  --go_out=plugins=grpc:src/ src/echo/echo.proto

#RUN GRPC_HEALTH_PROBE_VERSION=v0.2.0 && \
#    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
#    chmod +x /bin/grpc_health_probe

RUN export GOBIN=/app/bin && go install src/grpc_server.go
RUN export GOBIN=/app/bin && go install src/grpc_client.go

FROM gcr.io/distroless/base
COPY --from=build /app/bin /
COPY --from=build /app/xds_bootstrap.json /

EXPOSE 50051

#ENTRYPOINT ["grpc_server", "--grpcport", ":50051"]
#ENTRYPOINT ["grpc_client", "--host",  "server.domain.com:50051"]