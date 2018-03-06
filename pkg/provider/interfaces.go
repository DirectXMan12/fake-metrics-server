package provider

import (
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MetricsProvider provides metrics for pods and nodes
type MetricsProvider interface {
	ForPod(pod *v1.Pod) (info PodMetricInfo, found bool)
	ForNode(Node *v1.Node) (info NodeMetricInfo, found bool)
}

type MetricInfo interface {
	Timestamp() time.Time
	Window() time.Duration
}

type MetricValues interface {
	CPU() *resource.Quantity
	Memory() *resource.Quantity
}

type PodMetricInfo interface {
	MetricInfo
	ForContainer(*v1.Container) (vals MetricValues, found bool)
}

type NodeMetricInfo interface {
	MetricInfo
	MetricValues
}
