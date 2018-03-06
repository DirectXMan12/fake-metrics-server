package cmd

import (
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"
	"k8s.io/metrics/pkg/apis/metrics/install"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/directxman12/fake-metrics-server/cmd/options"
	"github.com/directxman12/fake-metrics-server/pkg/provider"
	nodemetricsstorage "github.com/directxman12/fake-metrics-server/pkg/storage/nodemetrics"
	podmetricsstorage "github.com/directxman12/fake-metrics-server/pkg/storage/podmetrics"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

func installMetricsAPIs(s *options.FakeMetricsServerRunOptions, g *genericapiserver.GenericAPIServer,
	metricsProvider provider.MetricsProvider, nodeLister v1listers.NodeLister, podLister v1listers.PodLister) {
	install.Install(groupFactoryRegistry, registry, Scheme)

	// we need to add the options to empty v1
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Group: "", Version: "v1"})

	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(metrics.GroupName, registry, Scheme, metav1.ParameterCodec, Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = v1beta1.SchemeGroupVersion

	nodemetricsStorage := nodemetricsstorage.NewStorage(metrics.Resource("nodemetrics"), metricsProvider, nodeLister)
	podmetricsStorage := podmetricsstorage.NewStorage(metrics.Resource("podmetrics"), metricsProvider, podLister)
	heapsterResources := map[string]rest.Storage{
		"nodes": nodemetricsStorage,
		"pods":  podmetricsStorage,
	}
	apiGroupInfo.VersionedResourcesStorageMap[v1beta1.SchemeGroupVersion.Version] = heapsterResources

	if err := g.InstallAPIGroup(&apiGroupInfo); err != nil {
		glog.Fatalf("Error in registering group versions: %v", err)
	}
}
