package proxy

import (
	"context"
	"sync"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	log "github.com/sirupsen/logrus"
)

type Callbacks struct {
	Signal   chan struct{}
	Fetches  int
	Requests int
	Mu       sync.Mutex
}

func (cb *Callbacks) Report() {
	cb.Mu.Lock()
	defer cb.Mu.Unlock()
	log.WithFields(log.Fields{"Fetches": cb.Fetches, "Requests": cb.Requests}).Info("Callbacks Report")
}

func (cb *Callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Infof("OnStreamOpen %d open for Type [%s]", id, typ)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64) {
	log.Infof("OnStreamClosed %d closed", id)
}

func (cb *Callbacks) OnStreamRequest(id int64, r *discovery.DiscoveryRequest) error {
	log.Infof("OnStreamRequest %d  Request[%v]", id, r.TypeUrl)
	cb.Mu.Lock()
	defer cb.Mu.Unlock()
	cb.Requests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnStreamResponse(id int64, req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
	log.Infof("OnStreamResponse... %d   Request [%v],  Response[%v]", id, req.TypeUrl, resp.TypeUrl)
	cb.Report()
}

func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *discovery.DiscoveryRequest) error {
	log.Infof("OnFetchRequest... Request [%v]", req.TypeUrl)
	cb.Mu.Lock()
	defer cb.Mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchResponse(req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
	log.Infof("OnFetchResponse... Resquest[%v],  Response[%v]", req.TypeUrl, resp.TypeUrl)
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64) {
	log.Infof("OnDeltaStreamClosed... %v", id)
}

func (cb *Callbacks) OnDeltaStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Infof("OnDeltaStreamOpen... %v  of type %s", id, typ)
	return nil
}

func (cb *Callbacks) OnStreamDeltaRequest(i int64, request *discovery.DeltaDiscoveryRequest) error {
	log.Infof("OnStreamDeltaRequest... %v  of type %s", i, request)
	return nil
}

func (cb *Callbacks) OnStreamDeltaResponse(i int64, request *discovery.DeltaDiscoveryRequest, response *discovery.DeltaDiscoveryResponse) {
	log.Infof("OnStreamDeltaResponse... %v  of type %s", i, request)
}
