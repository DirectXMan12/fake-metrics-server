package cmd

import (
	"fmt"
	"net"

	"github.com/directxman12/fake-metrics-server/cmd/options"
	"github.com/directxman12/fake-metrics-server/pkg/provider"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	v1listers "k8s.io/client-go/listers/core/v1"
)

const (
	msName = "Metrics Server"
)

type FakeMetricsServer struct {
	*genericapiserver.GenericAPIServer
	options    *options.FakeMetricsServerRunOptions
	metricsProvider provider.MetricsProvider
	nodeLister v1listers.NodeLister
}

// Run runs the specified APIServer. This should never exit.
func (h *FakeMetricsServer) RunServer() error {
	return h.PrepareRun().Run(wait.NeverStop)
}

func NewFakeMetricsServer(s *options.FakeMetricsServerRunOptions, metricsProvider provider.MetricsProvider,
	nodeLister v1listers.NodeLister, podLister v1listers.PodLister) (*FakeMetricsServer, error) {

	server, err := newAPIServer(s)
	if err != nil {
		return nil, err
	}

	installMetricsAPIs(s, server, metricsProvider, nodeLister, podLister)

	return &FakeMetricsServer{
		GenericAPIServer: server,
		options:          s,
		metricsProvider:  metricsProvider,     
		nodeLister:       nodeLister,
	}, nil
}

func newAPIServer(s *options.FakeMetricsServerRunOptions) (*genericapiserver.GenericAPIServer, error) {
	if err := s.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewConfig(Codecs)
	serverConfig.EnableMetrics = true

	if err := s.SecureServing.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	if !s.DisableAuthForTesting {
		if err := s.Authentication.ApplyTo(serverConfig); err != nil {
			return nil, err
		}
		if err := s.Authorization.ApplyTo(serverConfig); err != nil {
			return nil, err
		}
	}

	serverConfig.SwaggerConfig = genericapiserver.DefaultSwaggerConfig()

	return serverConfig.Complete().New(msName, genericapiserver.EmptyDelegate)
}
