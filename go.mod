module hello_xds

go 1.17

require (
	echo v0.0.0
	github.com/envoyproxy/go-control-plane v0.9.9
	github.com/golang/protobuf v1.5.2
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

require (
	cloud.google.com/go v0.34.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403 // indirect
	github.com/cncf/xds/go v0.0.0-20210805033703-aa0b78936158 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
)

replace echo => ./app/echo
