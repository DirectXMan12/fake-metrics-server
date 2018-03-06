package main

import (
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/directxman12/fake-metrics-server/cmd"
	"github.com/directxman12/fake-metrics-server/cmd/options"
	"github.com/directxman12/fake-metrics-server/pkg/provider"
)

func main() {
	opt := options.NewFakeMetricsServerRunOptions()
	opt.AddFlags(pflag.CommandLine)

	flag.InitFlags()

	logs.InitLogs()
	defer logs.FlushLogs()

	podLister, nodeLister := getListersOrDie()

	prov := provider.NewStaticProvider(resource.NewMilliQuantity(10000, resource.DecimalSI), resource.NewMilliQuantity(2, resource.DecimalSI))

	// Run API server
	server, err := cmd.NewFakeMetricsServer(opt, prov, nodeLister, podLister)
	if err != nil {
		glog.Fatalf("Could not create the API server: %v", err)
	}

	glog.Infof("Starting metrics.k8s.io API server...")
	glog.Fatal(server.RunServer())
}

func getListersOrDie() (v1listers.PodLister, v1listers.NodeLister) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to get Kubernetes client config: %v", err)
	}
	kubeClient := kubeclient.NewForConfigOrDie(kubeConfig)

	podLister, err := getPodLister(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create podLister: %v", err)
	}
	nodeLister, err := getNodeLister(kubeClient)
	if err != nil {
		glog.Fatalf("Failed to create nodeLister: %v", err)
	}
	return podLister, nodeLister
}

func getPodLister(kubeClient *kubeclient.Clientset) (v1listers.PodLister, error) {
	lw := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "pods", corev1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1listers.NewPodLister(store)
	reflector := cache.NewReflector(lw, &corev1.Pod{}, store, time.Hour)
	go reflector.Run(wait.NeverStop)
	return podLister, nil
}

func getNodeLister(kubeClient *kubeclient.Clientset) (v1listers.NodeLister, error) {
	lw := cache.NewListWatchFromClient(kubeClient.Core().RESTClient(), "nodes", corev1.NamespaceAll, fields.Everything())
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	nodeLister := v1listers.NewNodeLister(store)
	reflector := cache.NewReflector(lw, &corev1.Node{}, store, time.Hour)
	go reflector.Run(wait.NeverStop)

	return nodeLister, nil
}
