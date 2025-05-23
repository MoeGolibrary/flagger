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

package controller

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/rest"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/fluxcd/flagger/pkg/canary"
	clientset "github.com/fluxcd/flagger/pkg/client/clientset/versioned"
	flaggerscheme "github.com/fluxcd/flagger/pkg/client/clientset/versioned/scheme"
	flaggerinformers "github.com/fluxcd/flagger/pkg/client/informers/externalversions/flagger/v1beta1"
	"github.com/fluxcd/flagger/pkg/metrics"
	"github.com/fluxcd/flagger/pkg/metrics/observers"
	"github.com/fluxcd/flagger/pkg/notifier"
	"github.com/fluxcd/flagger/pkg/router"
	knative "knative.dev/serving/pkg/client/clientset/versioned"
)

const controllerAgentName = "flagger"

// Controller is managing the canary objects and schedules canary deployments
type Controller struct {
	kubeConfig           *rest.Config
	kubeClient           kubernetes.Interface
	knativeClient        knative.Interface
	flaggerClient        clientset.Interface
	flaggerInformers     Informers
	flaggerSynced        cache.InformerSynced
	flaggerWindow        time.Duration
	workqueue            workqueue.RateLimitingInterface
	eventRecorder        record.EventRecorder
	logger               *zap.SugaredLogger
	canaries             *sync.Map
	jobs                 map[string]CanaryJob
	recorder             metrics.Recorder
	notifier             notifier.Interface
	canaryFactory        *canary.Factory
	routerFactory        *router.Factory
	observerFactory      *observers.Factory
	meshProvider         string
	eventWebhook         string
	clusterName          string
	noCrossNamespaceRefs bool
}

type Informers struct {
	CanaryInformer flaggerinformers.CanaryInformer
	MetricInformer flaggerinformers.MetricTemplateInformer
	AlertInformer  flaggerinformers.AlertProviderInformer
}

func NewController(
	kubeClient kubernetes.Interface,
	knativeClient knative.Interface,
	flaggerClient clientset.Interface,
	flaggerInformers Informers,
	flaggerWindow time.Duration,
	logger *zap.SugaredLogger,
	notifier notifier.Interface,
	canaryFactory *canary.Factory,
	routerFactory *router.Factory,
	observerFactory *observers.Factory,
	meshProvider string,
	version string,
	eventWebhook string,
	clusterName string,
	noCrossNamespaceRefs bool,
	kubeConfig *rest.Config,
) *Controller {
	logger.Debug("Creating event broadcaster")
	flaggerscheme.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Named("event-broadcaster").Debugf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeClient.CoreV1().Events(""),
	})
	eventRecorder := eventBroadcaster.NewRecorder(
		scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	recorder := metrics.NewRecorder(controllerAgentName, true)
	recorder.SetInfo(version, meshProvider)

	ctrl := &Controller{
		kubeConfig:           kubeConfig,
		kubeClient:           kubeClient,
		knativeClient:        knativeClient,
		flaggerClient:        flaggerClient,
		flaggerInformers:     flaggerInformers,
		flaggerSynced:        flaggerInformers.CanaryInformer.Informer().HasSynced,
		workqueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerAgentName),
		eventRecorder:        eventRecorder,
		logger:               logger,
		canaries:             new(sync.Map),
		jobs:                 map[string]CanaryJob{},
		flaggerWindow:        flaggerWindow,
		observerFactory:      observerFactory,
		recorder:             recorder,
		notifier:             notifier,
		canaryFactory:        canaryFactory,
		routerFactory:        routerFactory,
		meshProvider:         meshProvider,
		eventWebhook:         eventWebhook,
		clusterName:          clusterName,
		noCrossNamespaceRefs: noCrossNamespaceRefs,
	}

	flaggerInformers.CanaryInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.enqueue,
		UpdateFunc: func(old, new interface{}) {
			oldCanary, ok := checkCustomResourceType(old, logger)
			if !ok {
				return
			}
			newCanary, ok := checkCustomResourceType(new, logger)
			if !ok {
				return
			}

			if diff := cmp.Diff(newCanary.Spec, oldCanary.Spec); diff != "" {
				ctrl.logger.Debugf("Diff detected %s.%s %s", oldCanary.Name, oldCanary.Namespace, diff)

				// warn about routing conflicts when service name changes
				if oldCanary.Spec.Service.Name != "" && oldCanary.Spec.Service.Name != newCanary.Spec.Service.Name {
					ctrl.logger.With("canary", fmt.Sprintf("%s.%s", oldCanary.Name, oldCanary.Namespace)).
						Warnf("The service name changed to %s, remove %s objects to avoid routing conflicts",
							newCanary.Spec.Service.Name, oldCanary.Spec.Service.Name)
				}

				ctrl.enqueue(new)
			} else if !newCanary.DeletionTimestamp.IsZero() && hasFinalizer(&newCanary) ||
				!hasFinalizer(&newCanary) && newCanary.Spec.RevertOnDeletion {
				// If this was marked for deletion and has finalizers enqueue for finalizing or
				// if this canary doesn't have finalizers and RevertOnDeletion is true updated speck enqueue
				ctrl.enqueue(new)
			}

			// If canary no longer desires reverting, finalizers should be removed
			if oldCanary.Spec.RevertOnDeletion && !newCanary.Spec.RevertOnDeletion {
				ctrl.logger.Infof("%s.%s opting out, deleting finalizers", newCanary.Name, newCanary.Namespace)
				err := ctrl.removeFinalizer(&newCanary)
				if err != nil {
					ctrl.logger.Warnf("Failed to remove finalizers for %s.%s: %v", oldCanary.Name, oldCanary.Namespace, err)
					return
				}
			}
		},
		DeleteFunc: func(old interface{}) {
			r, ok := checkCustomResourceType(old, logger)
			if ok {
				ctrl.logger.Infof("Deleting %s.%s from cache", r.Name, r.Namespace)
				ctrl.canaries.Delete(fmt.Sprintf("%s.%s", r.Name, r.Namespace))
			}
		},
	})

	return ctrl
}

// Run starts the K8s workers and the canary scheduler
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.logger.Info("Starting operator")

	for i := 0; i < threadiness; i++ {
		go wait.Until(func() {
			for c.processNextWorkItem() {
			}
		}, time.Second, stopCh)
	}

	c.logger.Info("Started operator workers")

	tickChan := time.NewTicker(c.flaggerWindow).C
	for {
		select {
		case <-tickChan:
			c.scheduleCanaries()
		case <-stopCh:
			c.logger.Info("Shutting down operator workers")
			return nil
		}
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Canary resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %w", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	cd, err := c.flaggerInformers.CanaryInformer.Lister().Canaries(namespace).Get(name)
	if errors.IsNotFound(err) {
		utilruntime.HandleError(fmt.Errorf("%s in work queue no longer exists", key))
		return nil
	}

	if err := c.verifyCanary(cd); err != nil {
		return fmt.Errorf("invalid canary spec: %s", err)
	}

	// Finalize if canary has been marked for deletion and revert is desired
	if cd.Spec.RevertOnDeletion && cd.ObjectMeta.DeletionTimestamp != nil {
		// If finalizers have been previously removed proceed
		if !hasFinalizer(cd) {
			c.logger.Infof("Canary %s.%s has been finalized", cd.Name, cd.Namespace)
			return nil
		}

		if cd.Status.Phase != flaggerv1.CanaryPhaseTerminated {
			if err := c.finalize(cd); err != nil {
				c.logger.With("canary", fmt.Sprintf("%s.%s", cd.Name, cd.Namespace)).
					Errorf("Unable to finalize canary: %v", err)
				return fmt.Errorf("unable to finalize to canary %s.%s error: %w", cd.Name, cd.Namespace, err)
			}
		}

		// Remove finalizer from Canary
		if err := c.removeFinalizer(cd); err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", cd.Name, cd.Namespace)).
				Errorf("Unable to remove finalizer for canary %s.%s error: %v", cd.Name, cd.Namespace, err)
			return fmt.Errorf("unable to remove finalizer for canary %s.%s: %w", cd.Name, cd.Namespace, err)
		}

		// record event
		c.recordEventInfof(cd, "Terminated canary %s.%s", cd.Name, cd.Namespace)

		c.logger.Infof("Canary %s.%s has been successfully processed and marked for deletion", cd.Name, cd.Namespace)
		return nil
	}

	// set status condition for new canaries
	if cd.Status.Conditions == nil {
		if err := c.setPhaseInitializing(cd); err != nil {
			c.logger.Errorf("%s unable to set initializing status: %v", key, err)
			return fmt.Errorf("%s initializing error: %w", key, err)
		}
	}

	c.canaries.Store(fmt.Sprintf("%s.%s", cd.Name, cd.Namespace), cd)

	// If opt in for revertOnDeletion add finalizer if not present
	if cd.Spec.RevertOnDeletion && !hasFinalizer(cd) {
		if err := c.addFinalizer(cd); err != nil {
			return fmt.Errorf("unable to add finalizer to canary %s.%s: %w", cd.Name, cd.Namespace, err)
		}

	}
	c.logger.Infof("Synced %s", key)

	return nil
}

func (c *Controller) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) verifyCanary(canary *flaggerv1.Canary) error {
	if c.noCrossNamespaceRefs {
		if err := verifyNoCrossNamespaceRefs(canary); err != nil {
			return err
		}
	}
	if err := verifyKnativeCanary(canary); err != nil {
		return err
	}
	if err := verifySessionAffinity(canary); err != nil {
		return err
	}

	return nil
}

func verifyNoCrossNamespaceRefs(canary *flaggerv1.Canary) error {
	if canary.Spec.UpstreamRef != nil && canary.Spec.UpstreamRef.Namespace != canary.Namespace {
		return fmt.Errorf("can't access gloo upstream %s.%s, cross-namespace references are blocked", canary.Spec.UpstreamRef.Name, canary.Spec.UpstreamRef.Namespace)
	}
	if canary.Spec.Analysis != nil {
		for _, metric := range canary.Spec.Analysis.Metrics {
			if metric.TemplateRef != nil && metric.TemplateRef.Namespace != canary.Namespace {
				return fmt.Errorf("can't access metric template %s.%s, cross-namespace references are blocked", metric.TemplateRef.Name, metric.TemplateRef.Namespace)
			}
		}
		for _, alert := range canary.Spec.Analysis.Alerts {
			if alert.ProviderRef.Namespace != canary.Namespace {
				return fmt.Errorf("can't access alert provider %s.%s, cross-namespace references are blocked", alert.ProviderRef.Name, alert.ProviderRef.Namespace)
			}
		}
	}
	return nil
}

func verifyKnativeCanary(canary *flaggerv1.Canary) error {
	if canary.Spec.TargetRef.IsKnativeService() != (canary.Spec.Provider == flaggerv1.KnativeProvider) {
		if canary.Spec.TargetRef.IsKnativeService() {
			return fmt.Errorf("can't use %s provider with Knative Service as target", canary.Spec.Provider)
		}
		return fmt.Errorf("can't use %s/%s as target if provider is set to knative",
			canary.Spec.TargetRef.APIVersion, canary.Spec.TargetRef.Kind)
	}

	if canary.Spec.TargetRef.IsKnativeService() {
		if canary.Spec.AutoscalerRef != nil {
			return fmt.Errorf("can't use autoscaler with Knative Service")
		}
	}

	return nil
}

func verifySessionAffinity(canary *flaggerv1.Canary) error {
	if canary.Spec.Analysis.SessionAffinity != nil {
		if canary.Spec.Analysis.SessionAffinity.CookieName == canary.Spec.Analysis.SessionAffinity.PrimaryCookieName {
			return fmt.Errorf("can't use the same cookie name for both primary and cookie name; please update them to be different")
		}
	}

	return nil
}

func checkCustomResourceType(obj interface{}, logger *zap.SugaredLogger) (flaggerv1.Canary, bool) {
	var roll *flaggerv1.Canary
	var ok bool
	if roll, ok = obj.(*flaggerv1.Canary); !ok {
		logger.Errorf("Event watch received an invalid object: %#v", obj)
		return flaggerv1.Canary{}, false
	}
	return *roll, true
}

func int32p(i int32) *int32 {
	return &i
}
