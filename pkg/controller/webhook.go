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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

func callWebhook(webhook string, payload interface{}, timeout string, retries int) error {
	_, err := callWebhookWithResponse(webhook, payload, timeout, retries)
	return err
}

func callWebhookWithResponse(webhook string, payload interface{}, timeout string, retries int) ([]byte, error) {
	payloadBin, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	hook, err := url.Parse(webhook)
	if err != nil {
		return nil, err
	}

	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = retries
	httpClient.Logger = nil

	req, err := retryablehttp.NewRequest("POST", hook.String(), bytes.NewBuffer(payloadBin))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if timeout == "" {
		timeout = "10s"
	}
	t, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, err
	}

	httpClient.HTTPClient.Timeout = t

	r, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %s", err.Error())
	}

	if r.StatusCode > 202 {
		return b, errors.New(string(b))
	}

	return b, nil
}

// CallWebhook does a HTTP POST to an external service and
// returns an error if the response status code is non-2xx
func CallWebhook(r flaggerv1.Canary, phase flaggerv1.CanaryPhase, w flaggerv1.CanaryWebhook) error {
	t := time.Now()

	payload := flaggerv1.CanaryWebhookPayload{
		Name:          r.Name,
		Namespace:     r.Namespace,
		Phase:         phase,
		Checksum:      r.CanaryChecksum(),
		BuildId:       r.Status.LastBuildId,
		Type:          w.Type,
		FailedChecks:  r.Status.FailedChecks,
		CanaryWeight:  r.Status.CanaryWeight,
		Iterations:    r.Status.Iterations,
		RemainingTime: r.GetRemainingTime(),
		Metadata: map[string]string{
			"timestamp":        strconv.FormatInt(t.UnixNano()/1000000, 10),
			"phase":            string(r.Status.Phase),
			"failedChecks":     strconv.Itoa(r.Status.FailedChecks),
			"canaryWeight":     strconv.Itoa(r.Status.CanaryWeight),
			"iterations":       strconv.Itoa(r.Status.Iterations),
			"lastBuildId":      r.Status.LastBuildId,
			"lastAppliedSpec":  r.Status.LastAppliedSpec,
			"lastPromotedSpec": r.Status.LastPromotedSpec,
		},
	}

	if w.Metadata != nil {
		for key, value := range *w.Metadata {
			if _, ok := payload.Metadata[key]; ok {
				continue
			}
			payload.Metadata[key] = value
		}
	}

	if len(w.Timeout) < 2 {
		w.Timeout = "10s"
	}

	return callWebhook(w.URL, payload, w.Timeout, w.Retries)
}

func CallEventWebhook(r *flaggerv1.Canary, w flaggerv1.CanaryWebhook, message, eventtype string) error {
	t := time.Now()

	payload := flaggerv1.CanaryWebhookPayload{
		Name:          r.Name,
		Namespace:     r.Namespace,
		Phase:         r.Status.Phase,
		Checksum:      r.CanaryChecksum(),
		BuildId:       r.Status.LastBuildId,
		Type:          w.Type,
		FailedChecks:  r.Status.FailedChecks,
		CanaryWeight:  r.Status.CanaryWeight,
		Iterations:    r.Status.Iterations,
		RemainingTime: r.GetRemainingTime(),
		Metadata: map[string]string{
			"eventMessage":     message,
			"eventType":        eventtype,
			"timestamp":        strconv.FormatInt(t.UnixNano()/1000000, 10),
			"phase":            string(r.Status.Phase),
			"failedChecks":     strconv.Itoa(r.Status.FailedChecks),
			"canaryWeight":     strconv.Itoa(r.Status.CanaryWeight),
			"iterations":       strconv.Itoa(r.Status.Iterations),
			"lastBuildId":      r.Status.LastBuildId,
			"lastAppliedSpec":  r.Status.LastAppliedSpec,
			"lastPromotedSpec": r.Status.LastPromotedSpec,
		},
	}

	if w.Metadata != nil {
		for key, value := range *w.Metadata {
			if _, ok := payload.Metadata[key]; ok {
				continue
			}
			payload.Metadata[key] = value
		}
	}
	return callWebhook(w.URL, payload, "5s", w.Retries)
}
