// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"
	_ "k8s.io/metrics/pkg/apis/metrics/install"

	"github.com/directxman12/fake-metrics-server/pkg/provider"
)

type MetricStorage struct {
	groupResource schema.GroupResource
	metricsProvider provider.MetricsProvider
	podLister     v1listers.PodLister
}

var _ rest.KindProvider = &MetricStorage{}
var _ rest.Storage = &MetricStorage{}
var _ rest.Getter = &MetricStorage{}
var _ rest.Lister = &MetricStorage{}

func NewStorage(groupResource schema.GroupResource, metricsProvider provider.MetricsProvider, podLister v1listers.PodLister) *MetricStorage {
	return &MetricStorage{
		groupResource: groupResource,
		metricsProvider: metricsProvider,
		podLister:     podLister,
	}
}

// Storage interface
func (m *MetricStorage) New() runtime.Object {
	return &metrics.PodMetrics{}
}

// KindProvider interface
func (m *MetricStorage) Kind() string {
	return "PodMetrics"
}

// Lister interface
func (m *MetricStorage) NewList() runtime.Object {
	return &metrics.PodMetricsList{}
}

// Lister interface
func (m *MetricStorage) List(ctx genericapirequest.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	labelSelector := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		labelSelector = options.LabelSelector
	}
	namespace := genericapirequest.NamespaceValue(ctx)
	pods, err := m.podLister.Pods(namespace).List(labelSelector)
	if err != nil {
		errMsg := fmt.Errorf("Error while listing pods for selector %v: %v", labelSelector, err)
		glog.Error(errMsg)
		return &metrics.PodMetricsList{}, errMsg
	}

	res := metrics.PodMetricsList{}
	for _, pod := range pods {
		if podMetrics := m.getPodMetrics(pod); podMetrics != nil {
			res.Items = append(res.Items, *podMetrics)
		} else {
			glog.Infof("No metrics for pod %s/%s", pod.Namespace, pod.Name)
		}
	}
	return &res, nil
}

// Getter interface
func (m *MetricStorage) Get(ctx genericapirequest.Context, name string, opts *metav1.GetOptions) (runtime.Object, error) {
	namespace := genericapirequest.NamespaceValue(ctx)

	pod, err := m.podLister.Pods(namespace).Get(name)
	if err != nil {
		errMsg := fmt.Errorf("Error while getting pod %v: %v", name, err)
		glog.Error(errMsg)
		return &metrics.PodMetrics{}, errMsg
	}
	if pod == nil {
		return &metrics.PodMetrics{}, errors.NewNotFound(v1.Resource("pods"), fmt.Sprintf("%v/%v", namespace, name))
	}

	podMetrics := m.getPodMetrics(pod)
	if podMetrics == nil {
		return &metrics.PodMetrics{}, errors.NewNotFound(m.groupResource, fmt.Sprintf("%v/%v", namespace, name))
	}
	return podMetrics, nil
}

func (m *MetricStorage) getPodMetrics(pod *v1.Pod) *metrics.PodMetrics {
	info, found := m.metricsProvider.ForPod(pod)
	if !found {
		glog.Errorf("No metrics for pod %s/%s", pod.Namespace, pod.Name)
		return nil
	}

	res := &metrics.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:              pod.Name,
			Namespace:         pod.Namespace,
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Timestamp:  metav1.NewTime(info.Timestamp()),
		Window:     metav1.Duration{Duration: info.Window()},
		Containers: make([]metrics.ContainerMetrics, 0),
	}

	for _, c := range pod.Spec.Containers {
		vals, found := info.ForContainer(&c)
		if !found {
			glog.Errorf("No metrics for container %s in pod %s/%s", c.Name, pod.Namespace, pod.Name)
			return nil
		}
		res.Containers = append(res.Containers, metrics.ContainerMetrics{
			Name: c.Name,
			Usage: metrics.ResourceList{
				metrics.ResourceName(v1.ResourceCPU.String()): *vals.CPU(),
				metrics.ResourceName(v1.ResourceMemory.String()): *vals.Memory(),
			},
		})
	}

	return res
}
