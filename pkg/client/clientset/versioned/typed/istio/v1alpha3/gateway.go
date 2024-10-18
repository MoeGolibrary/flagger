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

package v1alpha3

import (
	"context"

	v1alpha3 "github.com/fluxcd/flagger/pkg/apis/istio/v1alpha3"
	scheme "github.com/fluxcd/flagger/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// GatewaysGetter has a method to return a GatewayInterface.
// A group's client should implement this interface.
type GatewaysGetter interface {
	Gateways(namespace string) GatewayInterface
}

// GatewayInterface has methods to work with Gateway resources.
type GatewayInterface interface {
	Create(ctx context.Context, gateway *v1alpha3.Gateway, opts v1.CreateOptions) (*v1alpha3.Gateway, error)
	Update(ctx context.Context, gateway *v1alpha3.Gateway, opts v1.UpdateOptions) (*v1alpha3.Gateway, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha3.Gateway, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha3.GatewayList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha3.Gateway, err error)
	GatewayExpansion
}

// gateways implements GatewayInterface
type gateways struct {
	*gentype.ClientWithList[*v1alpha3.Gateway, *v1alpha3.GatewayList]
}

// newGateways returns a Gateways
func newGateways(c *NetworkingV1alpha3Client, namespace string) *gateways {
	return &gateways{
		gentype.NewClientWithList[*v1alpha3.Gateway, *v1alpha3.GatewayList](
			"gateways",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha3.Gateway { return &v1alpha3.Gateway{} },
			func() *v1alpha3.GatewayList { return &v1alpha3.GatewayList{} }),
	}
}
