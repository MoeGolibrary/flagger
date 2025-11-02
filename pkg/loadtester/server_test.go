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

package loadtester

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"

	"github.com/stretchr/testify/assert"
)

func TestServer_HandleHealthz(t *testing.T) {
	req, _ := http.NewRequest("GET", "/heathz", nil)
	resp := httptest.NewRecorder()
	HandleHealthz(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "OK", resp.Body.String())
}

func TestServer_HandleNewBashTaskCmdExitZero(t *testing.T) {
	mocks := newServerFixture()
	resp := mocks.resp
	req := newJsonRequest("POST", "/", &flaggerv1.CanaryWebhookPayload{
		Metadata: map[string]string{
			"type": TaskTypeBash,
			"cmd":  "echo some-output-not-to-be-returned",
		},
	})
	HandleNewTask(mocks.logger, mocks.taskRunner, NewAuthorizer(nil))(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Body.String())
}

func TestServer_HandleNewBashTaskCmdExitZeroReturnCmdOutput(t *testing.T) {
	mocks := newServerFixture()
	resp := mocks.resp
	req := newJsonRequest("POST", "/", &flaggerv1.CanaryWebhookPayload{
		Metadata: map[string]string{
			"type":            TaskTypeBash,
			"cmd":             "echo some-output-to-be-returned",
			"returnCmdOutput": "true",
		},
	})
	HandleNewTask(mocks.logger, mocks.taskRunner, NewAuthorizer(nil))(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "some-output-to-be-returned\n", resp.Body.String())
}

func TestServer_HandleNewBashTaskCmdExitNonZero(t *testing.T) {
	mocks := newServerFixture()
	resp := mocks.resp
	req := newJsonRequest("POST", "/", &flaggerv1.CanaryWebhookPayload{
		Metadata: map[string]string{
			"type": TaskTypeBash,
			"cmd":  "false",
		},
	})

	HandleNewTask(mocks.logger, mocks.taskRunner, NewAuthorizer(nil))(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, "command false failed: : exit status 1", resp.Body.String())
}

func TestServer_HandleSetManualTrafficControl(t *testing.T) {
	mocks := newServerFixture()
	resp := httptest.NewRecorder()
	
	// Create a request with manual traffic control state
	manualState := flaggerv1.CanaryManualState{
		Weight:    intPtr(40),
		Paused:    true,
		Timestamp: "1234567890",
	}
	
	payload, _ := json.Marshal(manualState)
	req, _ := http.NewRequest("POST", "/traffic/", bytes.NewReader(payload))
	req.Header.Set("Canary-Name", "podinfo")
	req.Header.Set("Canary-Namespace", "default")
	
	HandleSetManualTrafficControl(mocks.logger, NewAuthorizer(nil))(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	
	// Parse the response to verify it matches the expected structure
	var response struct {
		ManualTrafficControl struct {
			Weight    float64 `json:"weight"`
			Paused    bool    `json:"paused"`
			Timestamp string  `json:"timestamp"`
		} `json:"manualTrafficControl"`
	}
	
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(40), response.ManualTrafficControl.Weight)
	assert.Equal(t, true, response.ManualTrafficControl.Paused)
	assert.Equal(t, "1234567890", response.ManualTrafficControl.Timestamp)
}

func TestServer_HandleGetManualTrafficControl(t *testing.T) {
	mocks := newServerFixture()
	resp := httptest.NewRecorder()
	
	// First set a manual state
	manualState := flaggerv1.CanaryManualState{
		Weight:    intPtr(60),
		Paused:    false,
		Timestamp: "0987654321",
	}
	storeKey := "default.podinfo"
	manualStateStore.SetState(storeKey, &manualState)
	
	// Create a request to get the manual traffic control state
	payload := flaggerv1.CanaryWebhookPayload{
		Name:      "podinfo",
		Namespace: "default",
	}
	
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/traffic/state", bytes.NewReader(payloadBytes))
	
	HandleGetManualTrafficControl(mocks.logger, NewAuthorizer(nil))(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	
	// Parse the response to verify it matches the expected structure
	var response struct {
		ManualTrafficControl struct {
			Weight    float64 `json:"weight"`
			Paused    bool    `json:"paused"`
			Timestamp string  `json:"timestamp"`
		} `json:"manualTrafficControl"`
	}
	
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(60), response.ManualTrafficControl.Weight)
	assert.Equal(t, false, response.ManualTrafficControl.Paused)
	assert.Equal(t, "0987654321", response.ManualTrafficControl.Timestamp)
}

func newJsonRequest(method string, url string, v interface{}) *http.Request {
	payload, _ := json.Marshal(v)
	req, _ := http.NewRequest(method, url, bytes.NewReader(payload))
	return req
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}