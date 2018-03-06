package provider

import (
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type staticProvider struct {
	podInfo PodMetricInfo
	nodeInfo NodeMetricInfo
}

func NewStaticProvider(cpu *resource.Quantity, mem *resource.Quantity) MetricsProvider {
	info := &staticInfo{
		cpu: cpu,
		memory: mem,
		window: 1*time.Minute,
	}
	return &staticProvider{
		podInfo: info,
		nodeInfo: info,
	}
}

func (p *staticProvider) ForPod(_ *v1.Pod) (info PodMetricInfo, found bool) {
	return p.podInfo, true
}

func (p *staticProvider) ForNode(_ *v1.Node) (info NodeMetricInfo, found bool) {
	return p.nodeInfo, true
}

type staticInfo struct {
	cpu *resource.Quantity
	memory *resource.Quantity

	window time.Duration
}

func (i *staticInfo) Timestamp() time.Time {
	return time.Now()
}

func (i *staticInfo) Window() time.Duration {
	return i.window
}

func (i *staticInfo) CPU() *resource.Quantity {
	return i.cpu
}

func (i *staticInfo) Memory() *resource.Quantity {
	return i.memory
}

func (i *staticInfo) ForContainer(_ *v1.Container) (MetricValues, bool) {
	return i, true
}
