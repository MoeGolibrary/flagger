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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha3

import (
	v1alpha3 "github.com/fluxcd/flagger/pkg/apis/istio/v1alpha3"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// WorkloadEntryLister helps list WorkloadEntries.
// All objects returned here must be treated as read-only.
type WorkloadEntryLister interface {
	// List lists all WorkloadEntries in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha3.WorkloadEntry, err error)
	// WorkloadEntries returns an object that can list and get WorkloadEntries.
	WorkloadEntries(namespace string) WorkloadEntryNamespaceLister
	WorkloadEntryListerExpansion
}

// workloadEntryLister implements the WorkloadEntryLister interface.
type workloadEntryLister struct {
	listers.ResourceIndexer[*v1alpha3.WorkloadEntry]
}

// NewWorkloadEntryLister returns a new WorkloadEntryLister.
func NewWorkloadEntryLister(indexer cache.Indexer) WorkloadEntryLister {
	return &workloadEntryLister{listers.New[*v1alpha3.WorkloadEntry](indexer, v1alpha3.Resource("workloadentry"))}
}

// WorkloadEntries returns an object that can list and get WorkloadEntries.
func (s *workloadEntryLister) WorkloadEntries(namespace string) WorkloadEntryNamespaceLister {
	return workloadEntryNamespaceLister{listers.NewNamespaced[*v1alpha3.WorkloadEntry](s.ResourceIndexer, namespace)}
}

// WorkloadEntryNamespaceLister helps list and get WorkloadEntries.
// All objects returned here must be treated as read-only.
type WorkloadEntryNamespaceLister interface {
	// List lists all WorkloadEntries in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha3.WorkloadEntry, err error)
	// Get retrieves the WorkloadEntry from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha3.WorkloadEntry, error)
	WorkloadEntryNamespaceListerExpansion
}

// workloadEntryNamespaceLister implements the WorkloadEntryNamespaceLister
// interface.
type workloadEntryNamespaceLister struct {
	listers.ResourceIndexer[*v1alpha3.WorkloadEntry]
}
