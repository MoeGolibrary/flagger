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
	clientset "github.com/fluxcd/flagger/pkg/client/clientset/versioned"
	apisixv2 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/apisix/v2"
	fakeapisixv2 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/apisix/v2/fake"
	appmeshv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/appmesh/v1beta1"
	fakeappmeshv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/appmesh/v1beta1/fake"
	appmeshv1beta2 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/appmesh/v1beta2"
	fakeappmeshv1beta2 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/appmesh/v1beta2/fake"
	flaggerv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/flagger/v1beta1"
	fakeflaggerv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/flagger/v1beta1/fake"
	gatewayapiv1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/gatewayapi/v1"
	fakegatewayapiv1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/gatewayapi/v1/fake"
	gatewayapiv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/gatewayapi/v1beta1"
	fakegatewayapiv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/gatewayapi/v1beta1/fake"
	networkingv1alpha3 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/istio/v1alpha3"
	fakenetworkingv1alpha3 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/istio/v1alpha3/fake"
	networkingv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/istio/v1beta1"
	fakenetworkingv1beta1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/istio/v1beta1/fake"
	kedav1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/keda/v1alpha1"
	fakekedav1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/keda/v1alpha1/fake"
	kumav1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/kuma/v1alpha1"
	fakekumav1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/kuma/v1alpha1/fake"
	projectcontourv1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/projectcontour/v1"
	fakeprojectcontourv1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/projectcontour/v1/fake"
	splitv1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/smi/v1alpha1"
	fakesplitv1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/smi/v1alpha1/fake"
	splitv1alpha2 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/smi/v1alpha2"
	fakesplitv1alpha2 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/smi/v1alpha2/fake"
	splitv1alpha3 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/smi/v1alpha3"
	fakesplitv1alpha3 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/smi/v1alpha3/fake"
	traefikv1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/traefik/v1alpha1"
	faketraefikv1alpha1 "github.com/fluxcd/flagger/pkg/client/clientset/versioned/typed/traefik/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any field management, validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
//
// DEPRECATED: NewClientset replaces this with support for field management, which significantly improves
// server side apply testing. NewClientset is only available when apply configurations are generated (e.g.
// via --with-applyconfig).
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{tracker: o}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
	tracker   testing.ObjectTracker
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

func (c *Clientset) Tracker() testing.ObjectTracker {
	return c.tracker
}

var (
	_ clientset.Interface = &Clientset{}
	_ testing.FakeClient  = &Clientset{}
)

// ApisixV2 retrieves the ApisixV2Client
func (c *Clientset) ApisixV2() apisixv2.ApisixV2Interface {
	return &fakeapisixv2.FakeApisixV2{Fake: &c.Fake}
}

// AppmeshV1beta1 retrieves the AppmeshV1beta1Client
func (c *Clientset) AppmeshV1beta1() appmeshv1beta1.AppmeshV1beta1Interface {
	return &fakeappmeshv1beta1.FakeAppmeshV1beta1{Fake: &c.Fake}
}

// AppmeshV1beta2 retrieves the AppmeshV1beta2Client
func (c *Clientset) AppmeshV1beta2() appmeshv1beta2.AppmeshV1beta2Interface {
	return &fakeappmeshv1beta2.FakeAppmeshV1beta2{Fake: &c.Fake}
}

// FlaggerV1beta1 retrieves the FlaggerV1beta1Client
func (c *Clientset) FlaggerV1beta1() flaggerv1beta1.FlaggerV1beta1Interface {
	return &fakeflaggerv1beta1.FakeFlaggerV1beta1{Fake: &c.Fake}
}

// GatewayapiV1 retrieves the GatewayapiV1Client
func (c *Clientset) GatewayapiV1() gatewayapiv1.GatewayapiV1Interface {
	return &fakegatewayapiv1.FakeGatewayapiV1{Fake: &c.Fake}
}

// GatewayapiV1beta1 retrieves the GatewayapiV1beta1Client
func (c *Clientset) GatewayapiV1beta1() gatewayapiv1beta1.GatewayapiV1beta1Interface {
	return &fakegatewayapiv1beta1.FakeGatewayapiV1beta1{Fake: &c.Fake}
}

// NetworkingV1alpha3 retrieves the NetworkingV1alpha3Client
func (c *Clientset) NetworkingV1alpha3() networkingv1alpha3.NetworkingV1alpha3Interface {
	return &fakenetworkingv1alpha3.FakeNetworkingV1alpha3{Fake: &c.Fake}
}

// NetworkingV1beta1 retrieves the NetworkingV1beta1Client
func (c *Clientset) NetworkingV1beta1() networkingv1beta1.NetworkingV1beta1Interface {
	return &fakenetworkingv1beta1.FakeNetworkingV1beta1{Fake: &c.Fake}
}

// KedaV1alpha1 retrieves the KedaV1alpha1Client
func (c *Clientset) KedaV1alpha1() kedav1alpha1.KedaV1alpha1Interface {
	return &fakekedav1alpha1.FakeKedaV1alpha1{Fake: &c.Fake}
}

// KumaV1alpha1 retrieves the KumaV1alpha1Client
func (c *Clientset) KumaV1alpha1() kumav1alpha1.KumaV1alpha1Interface {
	return &fakekumav1alpha1.FakeKumaV1alpha1{Fake: &c.Fake}
}

// ProjectcontourV1 retrieves the ProjectcontourV1Client
func (c *Clientset) ProjectcontourV1() projectcontourv1.ProjectcontourV1Interface {
	return &fakeprojectcontourv1.FakeProjectcontourV1{Fake: &c.Fake}
}

// SplitV1alpha1 retrieves the SplitV1alpha1Client
func (c *Clientset) SplitV1alpha1() splitv1alpha1.SplitV1alpha1Interface {
	return &fakesplitv1alpha1.FakeSplitV1alpha1{Fake: &c.Fake}
}

// SplitV1alpha2 retrieves the SplitV1alpha2Client
func (c *Clientset) SplitV1alpha2() splitv1alpha2.SplitV1alpha2Interface {
	return &fakesplitv1alpha2.FakeSplitV1alpha2{Fake: &c.Fake}
}

// SplitV1alpha3 retrieves the SplitV1alpha3Client
func (c *Clientset) SplitV1alpha3() splitv1alpha3.SplitV1alpha3Interface {
	return &fakesplitv1alpha3.FakeSplitV1alpha3{Fake: &c.Fake}
}

// TraefikV1alpha1 retrieves the TraefikV1alpha1Client
func (c *Clientset) TraefikV1alpha1() traefikv1alpha1.TraefikV1alpha1Interface {
	return &faketraefikv1alpha1.FakeTraefikV1alpha1{Fake: &c.Fake}
}
