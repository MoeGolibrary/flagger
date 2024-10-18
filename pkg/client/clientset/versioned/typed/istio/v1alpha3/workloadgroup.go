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

// WorkloadGroupsGetter has a method to return a WorkloadGroupInterface.
// A group's client should implement this interface.
type WorkloadGroupsGetter interface {
	WorkloadGroups(namespace string) WorkloadGroupInterface
}

// WorkloadGroupInterface has methods to work with WorkloadGroup resources.
type WorkloadGroupInterface interface {
	Create(ctx context.Context, workloadGroup *v1alpha3.WorkloadGroup, opts v1.CreateOptions) (*v1alpha3.WorkloadGroup, error)
	Update(ctx context.Context, workloadGroup *v1alpha3.WorkloadGroup, opts v1.UpdateOptions) (*v1alpha3.WorkloadGroup, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, workloadGroup *v1alpha3.WorkloadGroup, opts v1.UpdateOptions) (*v1alpha3.WorkloadGroup, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha3.WorkloadGroup, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha3.WorkloadGroupList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha3.WorkloadGroup, err error)
	WorkloadGroupExpansion
}

// workloadGroups implements WorkloadGroupInterface
type workloadGroups struct {
	*gentype.ClientWithList[*v1alpha3.WorkloadGroup, *v1alpha3.WorkloadGroupList]
}

// newWorkloadGroups returns a WorkloadGroups
func newWorkloadGroups(c *NetworkingV1alpha3Client, namespace string) *workloadGroups {
	return &workloadGroups{
		gentype.NewClientWithList[*v1alpha3.WorkloadGroup, *v1alpha3.WorkloadGroupList](
			"workloadgroups",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha3.WorkloadGroup { return &v1alpha3.WorkloadGroup{} },
			func() *v1alpha3.WorkloadGroupList { return &v1alpha3.WorkloadGroupList{} }),
	}
}
