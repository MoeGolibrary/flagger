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
	"context"
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/fluxcd/flagger/pkg/metrics/observers"
	"github.com/fluxcd/flagger/pkg/metrics/providers"
)

const (
	MetricsProviderServiceSuffix = ":service"
)

var rateLimiter = rate.NewLimiter(rate.Every(time.Second), 10)

// to be called during canary initialization
func (c *Controller) checkMetricProviderAvailability(canary *flaggerv1.Canary) error {
	for _, metric := range canary.GetAnalysis().Metrics {
		if metric.Name == "request-success-rate" || metric.Name == "request-duration" {
			observerFactory := c.observerFactory
			if canary.Spec.MetricsServer != "" {
				var err error
				observerFactory, err = observers.NewFactory(canary.Spec.MetricsServer)
				if err != nil {
					return fmt.Errorf("error building Prometheus client for %s %v", canary.Spec.MetricsServer, err)
				}
			}
			if ok, err := observerFactory.Client.IsOnline(); !ok || err != nil {
				return fmt.Errorf("prometheus not avaiable: %v", err)
			}
			continue
		}

		if metric.TemplateRef != nil {
			namespace := canary.Namespace
			if metric.TemplateRef.Namespace != canary.Namespace && metric.TemplateRef.Namespace != "" {
				namespace = metric.TemplateRef.Namespace
			}

			template, err := c.flaggerInformers.MetricInformer.Lister().MetricTemplates(namespace).Get(metric.TemplateRef.Name)
			if err != nil {
				return fmt.Errorf("metric template %s.%s error: %v", metric.TemplateRef.Name, namespace, err)
			}

			var credentials map[string][]byte
			if template.Spec.Provider.SecretRef != nil {
				secret, err := c.kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), template.Spec.Provider.SecretRef.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("metric template %s.%s secret %s error: %v",
						metric.TemplateRef.Name, namespace, template.Spec.Provider.SecretRef.Name, err)
				}
				credentials = secret.Data
			}

			factory := providers.Factory{}
			provider, err := factory.Provider(metric.Interval, template.Spec.Provider, credentials, c.kubeConfig, c.logger)
			if err != nil {
				return fmt.Errorf("metric template %s.%s provider %s error: %v",
					metric.TemplateRef.Name, namespace, template.Spec.Provider.Type, err)
			}

			if ok, err := provider.IsOnline(); !ok || err != nil {
				return fmt.Errorf("%v in metric template %s.%s not avaiable: %v", template.Spec.Provider.Type,
					template.Name, template.Namespace, err)
			}
		}
	}
	c.recordEventInfof(canary, "all the metrics providers are available!")
	return nil
}

//func (c *Controller) runBuiltinMetricChecks(canary *flaggerv1.Canary) bool {
//	// override the global provider if one is specified in the canary spec
//	var metricsProvider string
//	// set the metrics provider to Crossover Prometheus when Crossover is the mesh provider
//	// For example, `crossover` metrics provider should be used for `smi:crossover` mesh provider
//	if strings.Contains(c.meshProvider, "crossover") {
//		metricsProvider = "crossover"
//	} else {
//		metricsProvider = c.meshProvider
//	}
//
//	if canary.Spec.Provider != "" {
//		metricsProvider = canary.Spec.Provider
//
//		// set the metrics provider to Linkerd Prometheus when Linkerd is the default mesh provider
//		if strings.Contains(c.meshProvider, "linkerd") {
//			metricsProvider = "linkerd"
//		}
//	}
//	// set the metrics provider to query Prometheus for the canary Kubernetes service if the canary target is Service
//	if canary.Spec.TargetRef.Kind == "Service" {
//		metricsProvider = metricsProvider + MetricsProviderServiceSuffix
//	}
//
//	// create observer based on the mesh provider
//	observerFactory := c.observerFactory
//
//	// override the global metrics server if one is specified in the canary spec
//	if canary.Spec.MetricsServer != "" {
//		var err error
//		observerFactory, err = observers.NewFactory(canary.Spec.MetricsServer)
//		if err != nil {
//			c.recordEventErrorf(canary, "Error building Prometheus client for %s %v", canary.Spec.MetricsServer, err)
//			return false
//		}
//	}
//	observer := observerFactory.Observer(metricsProvider)
//
//	// run metrics checks
//	for _, metric := range canary.GetAnalysis().Metrics {
//		if metric.Interval == "" {
//			metric.Interval = canary.GetMetricInterval()
//		}
//
//		if metric.Name == "request-success-rate" {
//			val, err := observer.GetRequestSuccessRate(toMetricModel(canary, metric.Interval, metric.TemplateVariables))
//			if err != nil {
//				if errors.Is(err, providers.ErrNoValuesFound) {
//					c.recordEventWarningf(canary,
//						"Halt advancement no values found for %s metric %s probably %s.%s is not receiving traffic: %v",
//						metricsProvider, metric.Name, canary.Spec.TargetRef.Name, canary.Namespace, err)
//				} else {
//					c.recordEventErrorf(canary, "Prometheus query failed: %v", err)
//				}
//				return false
//			}
//			c.recorder.SetAnalysis(canary, metric.Name, val)
//			if metric.ThresholdRange != nil {
//				tr := *metric.ThresholdRange
//				if tr.Min != nil && val < *tr.Min {
//					c.recordEventWarningf(canary, "Halt %s.%s advancement success rate %.2f%% < %v%%",
//						canary.Name, canary.Namespace, val, *tr.Min)
//					return false
//				}
//				if tr.Max != nil && val > *tr.Max {
//					c.recordEventWarningf(canary, "Halt %s.%s advancement success rate %.2f%% > %v%%",
//						canary.Name, canary.Namespace, val, *tr.Max)
//					return false
//				}
//			} else if metric.Threshold > val {
//				c.recordEventWarningf(canary, "Halt %s.%s advancement success rate %.2f%% < %v%%",
//					canary.Name, canary.Namespace, val, metric.Threshold)
//				return false
//			}
//		}
//
//		if metric.Name == "request-duration" {
//			val, err := observer.GetRequestDuration(toMetricModel(canary, metric.Interval, metric.TemplateVariables))
//			if err != nil {
//				if errors.Is(err, providers.ErrNoValuesFound) {
//					c.recordEventWarningf(canary, "Halt advancement no values found for %s metric %s probably %s.%s is not receiving traffic",
//						metricsProvider, metric.Name, canary.Spec.TargetRef.Name, canary.Namespace)
//				} else {
//					c.recordEventErrorf(canary, "Prometheus query failed: %v", err)
//				}
//				return false
//			}
//			c.recorder.SetAnalysis(canary, metric.Name, val.Seconds())
//			if metric.ThresholdRange != nil {
//				tr := *metric.ThresholdRange
//				if tr.Min != nil && val < time.Duration(*tr.Min)*time.Millisecond {
//					c.recordEventWarningf(canary, "Halt %s.%s advancement request duration %v < %v",
//						canary.Name, canary.Namespace, val, time.Duration(*tr.Min)*time.Millisecond)
//					return false
//				}
//				if tr.Max != nil && val > time.Duration(*tr.Max)*time.Millisecond {
//					c.recordEventWarningf(canary, "Halt %s.%s advancement request duration %v > %v",
//						canary.Name, canary.Namespace, val, time.Duration(*tr.Max)*time.Millisecond)
//					return false
//				}
//			} else if val > time.Duration(metric.Threshold)*time.Millisecond {
//				c.recordEventWarningf(canary, "Halt %s.%s advancement request duration %v > %v",
//					canary.Name, canary.Namespace, val, time.Duration(metric.Threshold)*time.Millisecond)
//				return false
//			}
//		}
//
//		// in-line PromQL
//		if metric.Query != "" {
//			query, err := observers.RenderQuery(metric.Query, toMetricModel(canary, metric.Interval, metric.TemplateVariables))
//			val, err := observerFactory.Client.RunQuery(query)
//			if err != nil {
//				if errors.Is(err, providers.ErrNoValuesFound) {
//					c.recordEventWarningf(canary, "Halt advancement no values found for metric: %s",
//						metric.Name)
//				} else {
//					c.recordEventErrorf(canary, "Prometheus query failed for %s: %v", metric.Name, err)
//				}
//				return false
//			}
//			c.recorder.SetAnalysis(canary, metric.Name, val)
//			if metric.ThresholdRange != nil {
//				tr := *metric.ThresholdRange
//				if tr.Min != nil && val < *tr.Min {
//					c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f < %v",
//						canary.Name, canary.Namespace, metric.Name, val, *tr.Min)
//					return false
//				}
//				if tr.Max != nil && val > *tr.Max {
//					c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f > %v",
//						canary.Name, canary.Namespace, metric.Name, val, *tr.Max)
//					return false
//				}
//			} else if val > metric.Threshold {
//				c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f > %v",
//					canary.Name, canary.Namespace, metric.Name, val, metric.Threshold)
//				return false
//			}
//		}
//	}
//
//	return true
//}

func (c *Controller) runMetricChecks(canary *flaggerv1.Canary) (bool, error) {
	for _, metric := range canary.GetAnalysis().Metrics {
		if metric.TemplateRef != nil {
			namespace := canary.Namespace
			if metric.TemplateRef.Namespace != canary.Namespace && metric.TemplateRef.Namespace != "" {
				namespace = metric.TemplateRef.Namespace
			}

			template, err := c.flaggerInformers.MetricInformer.Lister().MetricTemplates(namespace).Get(metric.TemplateRef.Name)
			if err != nil {
				c.recordEventErrorf(canary, "Metric template %s.%s error: %v", metric.TemplateRef.Name, namespace, err)
				return false, err
			}

			var credentials map[string][]byte
			if template.Spec.Provider.SecretRef != nil {
				secret, err := c.kubeClient.CoreV1().Secrets(namespace).Get(context.TODO(), template.Spec.Provider.SecretRef.Name, metav1.GetOptions{})
				if err != nil {
					c.recordEventErrorf(canary, "Metric template %s.%s secret %s error: %v",
						metric.TemplateRef.Name, namespace, template.Spec.Provider.SecretRef.Name, err)
					return false, err
				}
				credentials = secret.Data
			}

			factory := providers.Factory{}
			provider, err := factory.Provider(metric.Interval, template.Spec.Provider, credentials, c.kubeConfig, c.logger)
			if err != nil {
				c.recordEventErrorf(canary, "Metric template %s.%s provider %s error: %v",
					metric.TemplateRef.Name, namespace, template.Spec.Provider.Type, err)
				return false, err
			}

			query, err := observers.RenderQuery(template.Spec.Query, toMetricModel(canary, metric.Interval, metric.TemplateVariables))
			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Debugf("Metric template %s.%s query: %s", metric.TemplateRef.Name, namespace, query)
			if err != nil {
				c.recordEventErrorf(canary, "Metric template %s.%s query render error: %v",
					metric.TemplateRef.Name, namespace, err)
				return false, err
			}

			val, err := c.runQuery(query, canary, provider)
			if err != nil {
				if errors.Is(err, providers.ErrSkipAnalysis) {
					c.recordEventWarningf(canary, "Skipping analysis for %s.%s: %v",
						canary.Name, canary.Namespace, err)
				} else if errors.Is(err, providers.ErrTooManyRequests) {
					c.recordEventWarningf(canary, "Too many requests %s %s.%s: %v",
						metric.Name, canary.Name, canary.Namespace, err)
				} else if errors.Is(err, providers.ErrNoValuesFound) {
					c.recordEventWarningf(canary, "Halt advancement no values found for custom metric: %s: %v",
						metric.Name, err)
				} else {
					c.recordEventErrorf(canary, "Metric query failed for %s: %v", metric.Name, err)
				}
				return false, err
			}

			c.recorder.SetAnalysis(canary, metric.Name, val)

			if metric.ThresholdRange != nil {
				tr := *metric.ThresholdRange
				if tr.Min != nil && val < *tr.Min {
					c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f < %v",
						canary.Name, canary.Namespace, metric.Name, val, *tr.Min)
					return false, err
				}
				if tr.Max != nil && val > *tr.Max {
					c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f > %v",
						canary.Name, canary.Namespace, metric.Name, val, *tr.Max)
					return false, err
				}
			} else if val > metric.Threshold {
				c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f > %v",
					canary.Name, canary.Namespace, metric.Name, val, metric.Threshold)
				return false, err
			}
		} else if metric.Name != "request-success-rate" && metric.Name != "request-duration" && metric.Query == "" {
			c.recordEventErrorf(canary, "Metric query failed for no usable metrics template and query were configured")
			return false, providers.ErrNoValuesFound
		}
	}

	return true, nil
}

func toMetricModel(r *flaggerv1.Canary, interval string, variables map[string]string) flaggerv1.MetricTemplateModel {
	service := r.Spec.TargetRef.Name
	if r.Spec.Service.Name != "" {
		service = r.Spec.Service.Name
	}
	ingress := r.Spec.TargetRef.Name
	if r.Spec.IngressRef != nil {
		ingress = r.Spec.IngressRef.Name
	}
	route := r.Spec.TargetRef.Name
	if r.Spec.RouteRef != nil {
		route = r.Spec.RouteRef.Name
	}
	return flaggerv1.MetricTemplateModel{
		Name:      r.Name,
		Namespace: r.Namespace,
		Target:    r.Spec.TargetRef.Name,
		Service:   service,
		Ingress:   ingress,
		Route:     route,
		Interval:  interval,
		Variables: variables,
	}
}

func (c *Controller) runQuery(query string, canary *flaggerv1.Canary, provider providers.Interface) (float64, error) {
	maxRetries := 3
	baseRetryDelay := 10 * time.Second
	maxRetryDelay := 1 * time.Minute // Set a maximum retry delay

	// Initialize random number generator
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	ctx := context.Background()

	for i := 0; i <= maxRetries; i++ {
		// Wait for the rate limiter to allow the request
		if err := rateLimiter.Wait(ctx); err != nil {
			return 0, fmt.Errorf("rate limiter wait error: %w", err)
		}
		val, err := provider.RunQuery(query)
		if err == nil {
			return val, nil
		}

		if errors.Is(err, providers.ErrTooManyRequests) || errors.Is(err, providers.ErrNoValuesFound) {
			// Use the Canary's Interval for sleep
			interval := canary.GetAnalysisInterval()
			if interval > baseRetryDelay {
				interval = baseRetryDelay
			}
			// Exponential backoff with jitter and max delay
			retryDelay := baseRetryDelay + time.Duration(ra.Intn(int(baseRetryDelay)*(i+1)))
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}

			// Sleep in intervals and check skip analysis
			for totalSleep := time.Duration(0); totalSleep < retryDelay; totalSleep += interval {
				if totalSleep > 0 {
					c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
						With("canary_name", canary.Name).
						With("canary_namespace", canary.Namespace).
						Debugf("Request Metrics error, try later, no: %d, interval: %v, totalSleep: %v, retryDelay: %v,", i, retryDelay, totalSleep, retryDelay)
					time.Sleep(interval)
				}
				if c.checkSkipAnalysis(canary) {
					return 0, providers.ErrSkipAnalysis
				}
			}
		} else {
			return 0, err
		}
	}
	return 0, providers.ErrTooManyRequests
}

func (c *Controller) checkSkipAnalysis(canary *flaggerv1.Canary) bool {
	cd, err := c.flaggerClient.FlaggerV1beta1().Canaries(canary.Namespace).Get(context.TODO(), canary.Name, metav1.GetOptions{})
	if err != nil {
		c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
			With("canary_name", canary.Name).
			With("canary_namespace", canary.Namespace).
			Errorf("Canary %s.%s not found", canary.Name, canary.Namespace)
		return false
	}
	if cd.SkipAnalysis() {
		c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
			With("canary_name", canary.Name).
			With("canary_namespace", canary.Namespace).
			Info("Skipping analysis")
		return true
	}
	return false
}
