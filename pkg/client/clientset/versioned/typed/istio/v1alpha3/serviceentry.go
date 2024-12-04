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

// ServiceEntriesGetter has a method to return a ServiceEntryInterface.
// A group's client should implement this interface.
type ServiceEntriesGetter interface {
	ServiceEntries(namespace string) ServiceEntryInterface
}

// ServiceEntryInterface has methods to work with ServiceEntry resources.
type ServiceEntryInterface interface {
	Create(ctx context.Context, serviceEntry *v1alpha3.ServiceEntry, opts v1.CreateOptions) (*v1alpha3.ServiceEntry, error)
	Update(ctx context.Context, serviceEntry *v1alpha3.ServiceEntry, opts v1.UpdateOptions) (*v1alpha3.ServiceEntry, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha3.ServiceEntry, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha3.ServiceEntryList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha3.ServiceEntry, err error)
	ServiceEntryExpansion
}

// serviceEntries implements ServiceEntryInterface
type serviceEntries struct {
	*gentype.ClientWithList[*v1alpha3.ServiceEntry, *v1alpha3.ServiceEntryList]
}

// newServiceEntries returns a ServiceEntries
func newServiceEntries(c *NetworkingV1alpha3Client, namespace string) *serviceEntries {
	return &serviceEntries{
		gentype.NewClientWithList[*v1alpha3.ServiceEntry, *v1alpha3.ServiceEntryList](
			"serviceentries",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha3.ServiceEntry { return &v1alpha3.ServiceEntry{} },
			func() *v1alpha3.ServiceEntryList { return &v1alpha3.ServiceEntryList{} }),
	}
}