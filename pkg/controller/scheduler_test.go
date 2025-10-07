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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

func TestMain(m *testing.M) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query()["query"][0] == "vector(1)" {
			// for IsOnline invoked during canary initialization
			w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1545905245.458,"1"]}]}}`))
			return
		}
		w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1545905245.458,"100"]}]}}`))
	}))

	testMetricsServerURL = ts.URL
	defer ts.Close()
	os.Exit(m.Run())
}

func TestNextStepWeight(t *testing.T) {
	mocks := newDeploymentFixture(nil)

	// Test case 1: Normal step weights
	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.StepWeight = 0 // Clear the default StepWeight
	canary.Spec.Analysis.StepWeights = []int{10, 30, 50}

	t.Logf("Testing normal step weights: %v", canary.Spec.Analysis.StepWeights)

	// At start (weight 0) should return first step (10)
	step := mocks.ctrl.nextStepWeight(canary, 0)
	t.Logf("nextStepWeight(canary, 0) = %d", step)
	assert.Equal(t, 10, step)

	// At first step (weight 10) should return difference to next step (30-10=20)
	step = mocks.ctrl.nextStepWeight(canary, 10)
	t.Logf("nextStepWeight(canary, 10) = %d", step)
	assert.Equal(t, 20, step)

	// At intermediate weight (weight 20) - not matching any step, should return max step (100-20=80)
	step = mocks.ctrl.nextStepWeight(canary, 20)
	t.Logf("nextStepWeight(canary, 20) = %d", step)
	assert.Equal(t, 80, step)

	// Test case 2: Overflow step weights
	canary2 := newDeploymentTestCanary()
	canary2.Spec.Analysis.StepWeight = 0                     // Clear the default StepWeight
	canary2.Spec.Analysis.StepWeights = []int{1, 2, 10, 200} // 200 > 100 (total)

	t.Logf("Testing overflow step weights: %v", canary2.Spec.Analysis.StepWeights)

	// At start (weight 0) should return first step (1)
	step = mocks.ctrl.nextStepWeight(canary2, 0)
	t.Logf("nextStepWeight(canary2, 0) = %d", step)
	assert.Equal(t, 1, step)

	// At first step (weight 1) should return difference to next step (2-1=1)
	step = mocks.ctrl.nextStepWeight(canary2, 1)
	t.Logf("nextStepWeight(canary2, 1) = %d", step)
	assert.Equal(t, 1, step)

	// At second step (weight 2) should return difference to next step (10-2=8)
	step = mocks.ctrl.nextStepWeight(canary2, 2)
	t.Logf("nextStepWeight(canary2, 2) = %d", step)
	assert.Equal(t, 8, step)

	// At last step (weight 10) should return remaining to total (100-10=90)
	step = mocks.ctrl.nextStepWeight(canary2, 10)
	t.Logf("nextStepWeight(canary2, 10) = %d", step)
	assert.Equal(t, 90, step)
}

func TestRunAnalysis(t *testing.T) {
	mocks := newDeploymentFixture(nil)

	// Create a simple canary with metrics
	canary := newDeploymentTestCanary()
	canary.Spec.Analysis = &flaggerv1.CanaryAnalysis{
		Threshold:  10,
		StepWeight: 10,
		Metrics: []flaggerv1.CanaryMetric{
			{
				Name:      "request-success-rate",
				Threshold: 99,
				Interval:  "1m",
			},
		},
	}

	// Initialize the canary
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing
	canary.Status.FailedChecks = 0

	// Run the analysis
	ok, err := mocks.ctrl.runAnalysis(canary)

	// With our mock setup, the metrics should pass
	require.NoError(t, err)
	assert.True(t, ok)
}
