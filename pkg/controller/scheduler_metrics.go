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
			provider, err := factory.Provider(metric.Interval, metric.HistoryWindow, template.Spec.Provider, credentials, c.kubeConfig, c.logger)
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
			provider, err := factory.Provider(metric.Interval, metric.HistoryWindow, template.Spec.Provider, credentials, c.kubeConfig, c.logger)
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

			val, err := c.ExecuteCurrentQuery(query, canary, provider)
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

			c.recorder.SetAnalysis(canary, metric, val)

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

			// todo: check primary vs canary

			// check change rate
			if metric.HistoryWindow != "" && metric.ChangeThresholdRange != nil {
				previousVal, err := c.GetPreviousMetricValue(query, canary, provider)
				if err != nil {
					if errors.Is(err, providers.ErrHistoricalWindowNotConfigured) {
						continue
					}
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

				changeRate := (val - previousVal) / previousVal

				// c.recorder.SetAnalysis(canary, metric, val)

				tr := *metric.ChangeThresholdRange
				if tr.Min != nil && changeRate < *tr.Min {
					c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f < %v",
						canary.Name, canary.Namespace, metric.Name, changeRate, *tr.Min)
					return false, err
				}
				if tr.Max != nil && changeRate > *tr.Max {
					c.recordEventWarningf(canary, "Halt %s.%s advancement %s %.2f > %v",
						canary.Name, canary.Namespace, metric.Name, changeRate, *tr.Max)
					return false, err
				}
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

func (c *Controller) ExecuteCurrentQuery(query string, canary *flaggerv1.Canary, provider providers.Interface) (float64, error) {
	maxRetries := 5
	baseRetryDelay := 3 * time.Second
	maxRetryDelay := 5 * time.Second // Set a maximum retry delay

	// Initialize random number generator
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	ctx := context.Background()

	for i := 0; i <= maxRetries; i++ {
		// Wait for the rate limiter to allow the request
		if err := rateLimiter.Wait(ctx); err != nil {
			return 0, fmt.Errorf("rate limiter wait error: %w", err)
		}
		val, err := provider.ExecuteCurrentQuery(query)
		if err == nil {
			return val, nil
		}

		if errors.Is(err, providers.ErrTooManyRequests) {
			// Use the Canary's Interval for sleep
			retryDelay := baseRetryDelay + time.Duration(ra.Intn(int(maxRetryDelay)))

			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Debugf("Request Metrics error, try later, no: %d, retryDelay: %v", i, retryDelay)
			time.Sleep(retryDelay)
			if c.checkSkipAnalysis(canary) {
				return 0, providers.ErrSkipAnalysis
			}
		} else {
			return 0, err
		}
	}
	return 0, providers.ErrTooManyRequests
}

func (c *Controller) GetPreviousMetricValue(query string, canary *flaggerv1.Canary, provider providers.Interface) (float64, error) {
	maxRetries := 5
	baseRetryDelay := 3 * time.Second
	maxRetryDelay := 5 * time.Second // Set a maximum retry delay

	// Initialize random number generator
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	ctx := context.Background()

	for i := 0; i <= maxRetries; i++ {
		// Wait for the rate limiter to allow the request
		if err := rateLimiter.Wait(ctx); err != nil {
			return 0, fmt.Errorf("rate limiter wait error: %w", err)
		}
		val, err := provider.GetPreviousMetricValue(query)
		if err == nil {
			return val, nil
		}

		if errors.Is(err, providers.ErrTooManyRequests) {
			// Use the Canary's Interval for sleep
			retryDelay := baseRetryDelay + time.Duration(ra.Intn(int(maxRetryDelay)))

			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Debugf("Request Metrics error, try later, no: %d, retryDelay: %v", i, retryDelay)
			time.Sleep(retryDelay)
			if c.checkSkipAnalysis(canary) {
				return 0, providers.ErrSkipAnalysis
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
