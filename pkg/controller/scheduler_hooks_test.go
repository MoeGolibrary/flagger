package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

// parseTrafficControlResponse was a helper function that existed in a previous version
// but is no longer needed in the current implementation

// TestSetManualTrafficControlState and related tests were for helper methods
// that existed in a previous version but are no longer needed in the current implementation

func TestRunManualTrafficControlHooks_Pause(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"weight\": 30, \"paused\": true, \"timestamp\": \"1234567890\"}"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.NoError(t, err)

	assert.NotNil(t, manualState)
	assert.Equal(t, 30, *manualState.Weight)
	assert.True(t, manualState.Paused)
}

func TestRunManualTrafficControlHooks_Resume(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"weight\": 30, \"paused\": false, \"timestamp\": \"1234567890\"}"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseWaiting

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.NoError(t, err)

	assert.NotNil(t, manualState)
	assert.Equal(t, 30, *manualState.Weight)
	assert.False(t, manualState.Paused)
}

func TestRunManualTrafficControlHooks_InvalidResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{invalid json"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.Error(t, err)
	assert.Nil(t, manualState)
}

func TestRunManualTrafficControlHooks_NoManualHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.NoError(t, err)
	assert.Nil(t, manualState)
}

func TestRunManualTrafficControlHooks_MultipleHooks(t *testing.T) {
	// First webhook returns success (resume), second returns pause
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"weight\": 30, \"paused\": false, \"timestamp\": \"1234567890\"}"))
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"weight\": 40, \"paused\": true, \"timestamp\": \"1234567891\"}"))
	}))
	defer ts2.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "manual-traffic-control-1",
			URL:  ts1.URL,
			Type: flaggerv1.ManualTrafficControlHook,
		},
		{
			Name: "manual-traffic-control-2",
			URL:  ts2.URL,
			Type: flaggerv1.ManualTrafficControlHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canaryObj)
	require.NoError(t, err)

	assert.NotNil(t, manualState)
	// First webhook should be used since it's successful
	assert.Equal(t, 30, *manualState.Weight)
	assert.False(t, manualState.Paused)
}

func TestRunManualTrafficControlHooks_WebhookError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("some error"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.Error(t, err)
	assert.Nil(t, manualState)
}

func newDeploymentTestCanaryWithManualHook(webhookURL string) *flaggerv1.Canary {
	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name:    "manual-traffic-control",
			URL:     webhookURL,
			Type:    flaggerv1.ManualTrafficControlHook,
			Retries: 0,
		},
	}
	return canary
}

func TestRunConfirmTrafficIncreaseHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmTrafficIncreaseHooks(canary)

	assert.True(t, shouldContinue)
}

func TestRunConfirmTrafficIncreaseHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "confirm-traffic-increase",
			URL:  ts.URL,
			Type: flaggerv1.ConfirmTrafficIncreaseHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmTrafficIncreaseHooks(canaryObj)

	assert.True(t, shouldContinue)
}

func TestRunConfirmTrafficIncreaseHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "confirm-traffic-increase",
			URL:  ts.URL,
			Type: flaggerv1.ConfirmTrafficIncreaseHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmTrafficIncreaseHooks(canaryObj)

	assert.False(t, shouldContinue)
}
