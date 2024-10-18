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

	v1alpha3 "github.com/fluxcd/flagger/pkg/apis/istio/v1alpha3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGateways implements GatewayInterface
type FakeGateways struct {
	Fake *FakeNetworkingV1alpha3
	ns   string
}

var gatewaysResource = v1alpha3.SchemeGroupVersion.WithResource("gateways")

var gatewaysKind = v1alpha3.SchemeGroupVersion.WithKind("Gateway")

// Get takes name of the gateway, and returns the corresponding gateway object, and an error if there is any.
func (c *FakeGateways) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha3.Gateway, err error) {
	emptyResult := &v1alpha3.Gateway{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(gatewaysResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha3.Gateway), err
}

// List takes label and field selectors, and returns the list of Gateways that match those selectors.
func (c *FakeGateways) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha3.GatewayList, err error) {
	emptyResult := &v1alpha3.GatewayList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(gatewaysResource, gatewaysKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha3.GatewayList{ListMeta: obj.(*v1alpha3.GatewayList).ListMeta}
	for _, item := range obj.(*v1alpha3.GatewayList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested gateways.
func (c *FakeGateways) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(gatewaysResource, c.ns, opts))

}

// Create takes the representation of a gateway and creates it.  Returns the server's representation of the gateway, and an error, if there is any.
func (c *FakeGateways) Create(ctx context.Context, gateway *v1alpha3.Gateway, opts v1.CreateOptions) (result *v1alpha3.Gateway, err error) {
	emptyResult := &v1alpha3.Gateway{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(gatewaysResource, c.ns, gateway, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha3.Gateway), err
}

// Update takes the representation of a gateway and updates it. Returns the server's representation of the gateway, and an error, if there is any.
func (c *FakeGateways) Update(ctx context.Context, gateway *v1alpha3.Gateway, opts v1.UpdateOptions) (result *v1alpha3.Gateway, err error) {
	emptyResult := &v1alpha3.Gateway{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(gatewaysResource, c.ns, gateway, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha3.Gateway), err
}

// Delete takes name of the gateway and deletes it. Returns an error if one occurs.
func (c *FakeGateways) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(gatewaysResource, c.ns, name, opts), &v1alpha3.Gateway{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGateways) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(gatewaysResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha3.GatewayList{})
	return err
}

// Patch applies the patch and returns the patched gateway.
func (c *FakeGateways) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha3.Gateway, err error) {
	emptyResult := &v1alpha3.Gateway{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(gatewaysResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha3.Gateway), err
}
