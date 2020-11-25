FROM golang:1.15.0-alpine3.12 as build

WORKDIR /xds
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/

ADD xds /xds

# COPY go-control-plane/go.mod go-control-plane/go.sum ./go-control-plane/
RUN go mod download
RUN export GOBIN=/xds/bin && go install main.go

FROM alpine:3.12.0
COPY --from=build /xds/bin /

ENTRYPOINT [ "./main" ]
