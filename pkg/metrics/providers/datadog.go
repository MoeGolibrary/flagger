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

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// https://docs.datadoghq.com/api/
const (
	datadogDefaultHost = "https://api.datadoghq.com"

	datadogMetricsQueryPath     = "/api/v1/query"
	datadogAPIKeyValidationPath = "/api/v1/validate"

	datadogKeysSecretKey = "datadog_keys"

	datadogAPIKeySecretKey = "datadog_api_key"
	datadogAPIKeyHeaderKey = "DD-API-KEY"

	datadogApplicationKeySecretKey = "datadog_application_key"
	datadogApplicationKeyHeaderKey = "DD-APPLICATION-KEY"

	datadogFromDeltaMultiplierOnMetricInterval = 10
)

// DatadogProvider executes datadog queries
type DatadogProvider struct {
	metricsQueryEndpoint     string
	apiKeyValidationEndpoint string
	timeout                  time.Duration
	apiKey                   string
	applicationKey           string
	keys                     []datadogKey
	fromDelta                int64
	history                  int64
	logger                   *zap.SugaredLogger
	ra                       *rand.Rand
}

type datadogKey struct {
	ApiKey         string `json:"api_key" yaml:"apiKey"`
	ApplicationKey string `json:"application_key" yaml:"applicationKey"`
}

type datadogResponse struct {
	Series []struct {
		Pointlist [][]float64 `json:"pointlist"`
	}
}

// NewDatadogProvider takes a canary spec, a provider spec and the credentials map, and
// returns a Datadog client ready to execute queries against the API
func NewDatadogProvider(metricInterval string,
	metricHistoryWindow string,
	provider flaggerv1.MetricTemplateProvider,
	credentials map[string][]byte,
	logger *zap.SugaredLogger) (*DatadogProvider, error) {

	address := provider.Address
	if address == "" {
		address = datadogDefaultHost
	}

	dd := DatadogProvider{
		timeout:                  5 * time.Second,
		metricsQueryEndpoint:     address + datadogMetricsQueryPath,
		apiKeyValidationEndpoint: address + datadogAPIKeyValidationPath,
		logger:                   logger,
		ra:                       rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if b, ok := credentials[datadogKeysSecretKey]; ok {
		if err := json.Unmarshal(b, &dd.keys); err != nil {
			logger.Error("error unmarshaling datadog keys", zap.Error(err))
		}
	}

	if b, ok := credentials[datadogAPIKeySecretKey]; ok {
		dd.apiKey = string(b)
	}
	if b, ok := credentials[datadogApplicationKeySecretKey]; ok {
		dd.applicationKey = string(b)
	}
	if dd.apiKey != "" && dd.applicationKey != "" {
		dd.keys = append(dd.keys, datadogKey{ApiKey: dd.apiKey, ApplicationKey: dd.applicationKey})
	}

	if len(dd.keys) == 0 {
		return nil, fmt.Errorf("no valid datadog keys found")
	}

	md, err := time.ParseDuration(metricInterval)
	if err != nil {
		return nil, fmt.Errorf("error parsing metric interval: %w", err)
	}

	dd.fromDelta = int64(datadogFromDeltaMultiplierOnMetricInterval * md.Seconds())

	if metricHistoryWindow != "" {
		historyWindow, err := time.ParseDuration(metricHistoryWindow)
		if err != nil {
			return nil, fmt.Errorf("error parsing metric history window: %w", err)
		}
		dd.history = int64(historyWindow.Seconds())

	}
	return &dd, nil
}

// ExecuteCurrentQuery executes the query for the current time window (now - fromDelta to now)
func (p *DatadogProvider) ExecuteCurrentQuery(query string) (float64, error) {
	now := time.Now().Unix()
	return p.runQueryTimeRange(query,
		strconv.FormatInt(now-p.fromDelta, 10),
		strconv.FormatInt(now, 10))
}

// GetPreviousMetricValue retrieves the metric value from the configured historical time window
func (p *DatadogProvider) GetPreviousMetricValue(query string) (float64, error) {
	if p.history == 0 {
		return 0, ErrHistoricalWindowNotConfigured
	}
	targetTime := time.Now().Unix() - p.history
	return p.runQueryTimeRange(query,
		strconv.FormatInt(targetTime-p.fromDelta, 10),
		strconv.FormatInt(targetTime, 10))
}

// RunQueryTimeRange executes the datadog query against DatadogProvider.metricsQueryEndpoint
func (p *DatadogProvider) runQueryTimeRange(query string, from, to string) (float64, error) {

	req, err := http.NewRequest("GET", p.metricsQueryEndpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("error http.NewRequest: %w", err)
	}

	// 随机从keys选择一个
	key := p.keys[p.ra.Intn(len(p.keys))]
	req.Header.Set(datadogAPIKeyHeaderKey, key.ApiKey)
	req.Header.Set(datadogApplicationKeyHeaderKey, key.ApplicationKey)

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("from", from)
	q.Add("to", to)
	req.URL.RawQuery = q.Encode()
	// Log the request details
	p.logger.Debugf("Request URL: %s", req.URL.String())
	p.logger.Debugf("Request Headers: %+v", req.Header)
	ctx, cancel := context.WithTimeout(req.Context(), p.timeout)
	defer cancel()
	r, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	p.logger.Debugf("Response Headers: %+v", r.Header)

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading body: %w", err)
	}
	// Log the response details
	p.logger.Debugf("Response Status: %s", r.Status)
	p.logger.Debugf("Response Body: %s", string(b))

	if r.StatusCode != http.StatusOK {
		if r.StatusCode == http.StatusTooManyRequests {
			return 0, ErrTooManyRequests
		}
		return 0, fmt.Errorf("error response: %s: %w", string(b), err)
	}

	var res datadogResponse
	if err := json.Unmarshal(b, &res); err != nil {
		return 0, fmt.Errorf("error unmarshaling result: %w, '%s'", err, string(b))
	}

	if len(res.Series) < 1 {
		return 0, fmt.Errorf("invalid response: %s: %w", string(b), ErrNoValuesFound)
	}

	pl := res.Series[0].Pointlist
	if len(pl) < 1 {
		return 0, fmt.Errorf("invalid response: %s: %w", string(b), ErrNoValuesFound)
	}

	vs := pl[len(pl)-1]
	if len(vs) < 1 {
		return 0, fmt.Errorf("invalid response: %s: %w", string(b), ErrNoValuesFound)
	}

	return vs[1], nil
}

// IsOnline calls the Datadog's validation endpoint with api keys
// and returns an error if the validation fails
func (p *DatadogProvider) IsOnline() (bool, error) {
	req, err := http.NewRequest("GET", p.apiKeyValidationEndpoint, nil)
	if err != nil {
		return false, fmt.Errorf("error http.NewRequest: %w", err)
	}

	// 随机从keys选择一个
	key := p.keys[p.ra.Intn(len(p.keys))]
	p.logger.Debugf("Datadog API Key: %s", key.ApiKey)
	req.Header.Set(datadogAPIKeyHeaderKey, key.ApiKey)
	req.Header.Set(datadogApplicationKeyHeaderKey, key.ApplicationKey)

	ctx, cancel := context.WithTimeout(req.Context(), p.timeout)
	defer cancel()
	r, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}

	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return false, fmt.Errorf("error reading body: %w", err)
	}

	if r.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error response: %s", string(b))
	}

	return true, nil
}
