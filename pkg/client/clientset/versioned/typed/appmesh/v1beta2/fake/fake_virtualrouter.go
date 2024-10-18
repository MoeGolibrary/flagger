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

	v1beta2 "github.com/fluxcd/flagger/pkg/apis/appmesh/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVirtualRouters implements VirtualRouterInterface
type FakeVirtualRouters struct {
	Fake *FakeAppmeshV1beta2
	ns   string
}

var virtualroutersResource = v1beta2.SchemeGroupVersion.WithResource("virtualrouters")

var virtualroutersKind = v1beta2.SchemeGroupVersion.WithKind("VirtualRouter")

// Get takes name of the virtualRouter, and returns the corresponding virtualRouter object, and an error if there is any.
func (c *FakeVirtualRouters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta2.VirtualRouter, err error) {
	emptyResult := &v1beta2.VirtualRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(virtualroutersResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta2.VirtualRouter), err
}

// List takes label and field selectors, and returns the list of VirtualRouters that match those selectors.
func (c *FakeVirtualRouters) List(ctx context.Context, opts v1.ListOptions) (result *v1beta2.VirtualRouterList, err error) {
	emptyResult := &v1beta2.VirtualRouterList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(virtualroutersResource, virtualroutersKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta2.VirtualRouterList{ListMeta: obj.(*v1beta2.VirtualRouterList).ListMeta}
	for _, item := range obj.(*v1beta2.VirtualRouterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested virtualRouters.
func (c *FakeVirtualRouters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(virtualroutersResource, c.ns, opts))

}

// Create takes the representation of a virtualRouter and creates it.  Returns the server's representation of the virtualRouter, and an error, if there is any.
func (c *FakeVirtualRouters) Create(ctx context.Context, virtualRouter *v1beta2.VirtualRouter, opts v1.CreateOptions) (result *v1beta2.VirtualRouter, err error) {
	emptyResult := &v1beta2.VirtualRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(virtualroutersResource, c.ns, virtualRouter, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta2.VirtualRouter), err
}

// Update takes the representation of a virtualRouter and updates it. Returns the server's representation of the virtualRouter, and an error, if there is any.
func (c *FakeVirtualRouters) Update(ctx context.Context, virtualRouter *v1beta2.VirtualRouter, opts v1.UpdateOptions) (result *v1beta2.VirtualRouter, err error) {
	emptyResult := &v1beta2.VirtualRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(virtualroutersResource, c.ns, virtualRouter, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta2.VirtualRouter), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVirtualRouters) UpdateStatus(ctx context.Context, virtualRouter *v1beta2.VirtualRouter, opts v1.UpdateOptions) (result *v1beta2.VirtualRouter, err error) {
	emptyResult := &v1beta2.VirtualRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(virtualroutersResource, "status", c.ns, virtualRouter, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta2.VirtualRouter), err
}

// Delete takes name of the virtualRouter and deletes it. Returns an error if one occurs.
func (c *FakeVirtualRouters) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(virtualroutersResource, c.ns, name, opts), &v1beta2.VirtualRouter{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVirtualRouters) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(virtualroutersResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta2.VirtualRouterList{})
	return err
}

// Patch applies the patch and returns the patched virtualRouter.
func (c *FakeVirtualRouters) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta2.VirtualRouter, err error) {
	emptyResult := &v1beta2.VirtualRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(virtualroutersResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta2.VirtualRouter), err
}
