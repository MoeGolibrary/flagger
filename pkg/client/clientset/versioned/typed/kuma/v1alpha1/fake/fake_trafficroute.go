/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/fluxcd/flagger/pkg/apis/kuma/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTrafficRoutes implements TrafficRouteInterface
type FakeTrafficRoutes struct {
	Fake *FakeKumaV1alpha1
}

var trafficroutesResource = v1alpha1.SchemeGroupVersion.WithResource("trafficroutes")

var trafficroutesKind = v1alpha1.SchemeGroupVersion.WithKind("TrafficRoute")

// Get takes name of the trafficRoute, and returns the corresponding trafficRoute object, and an error if there is any.
func (c *FakeTrafficRoutes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TrafficRoute, err error) {
	emptyResult := &v1alpha1.TrafficRoute{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(trafficroutesResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TrafficRoute), err
}

// List takes label and field selectors, and returns the list of TrafficRoutes that match those selectors.
func (c *FakeTrafficRoutes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TrafficRouteList, err error) {
	emptyResult := &v1alpha1.TrafficRouteList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(trafficroutesResource, trafficroutesKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.TrafficRouteList{ListMeta: obj.(*v1alpha1.TrafficRouteList).ListMeta}
	for _, item := range obj.(*v1alpha1.TrafficRouteList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested trafficRoutes.
func (c *FakeTrafficRoutes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(trafficroutesResource, opts))
}

// Create takes the representation of a trafficRoute and creates it.  Returns the server's representation of the trafficRoute, and an error, if there is any.
func (c *FakeTrafficRoutes) Create(ctx context.Context, trafficRoute *v1alpha1.TrafficRoute, opts v1.CreateOptions) (result *v1alpha1.TrafficRoute, err error) {
	emptyResult := &v1alpha1.TrafficRoute{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(trafficroutesResource, trafficRoute, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TrafficRoute), err
}

// Update takes the representation of a trafficRoute and updates it. Returns the server's representation of the trafficRoute, and an error, if there is any.
func (c *FakeTrafficRoutes) Update(ctx context.Context, trafficRoute *v1alpha1.TrafficRoute, opts v1.UpdateOptions) (result *v1alpha1.TrafficRoute, err error) {
	emptyResult := &v1alpha1.TrafficRoute{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(trafficroutesResource, trafficRoute, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TrafficRoute), err
}

// Delete takes name of the trafficRoute and deletes it. Returns an error if one occurs.
func (c *FakeTrafficRoutes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(trafficroutesResource, name, opts), &v1alpha1.TrafficRoute{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTrafficRoutes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(trafficroutesResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.TrafficRouteList{})
	return err
}

// Patch applies the patch and returns the patched trafficRoute.
func (c *FakeTrafficRoutes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TrafficRoute, err error) {
	emptyResult := &v1alpha1.TrafficRoute{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(trafficroutesResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TrafficRoute), err
}