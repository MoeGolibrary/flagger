package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

func TestParseTrafficControlResponse(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedRatio int
		expectedPause bool
	}{
		{
			name:          "valid pause with ratio 10",
			err:           fmt.Errorf("PAUSE:10"),
			expectedRatio: 10,
			expectedPause: true,
		},
		{
			name:          "valid pause with ratio 50",
			err:           fmt.Errorf("PAUSE:50"),
			expectedRatio: 50,
			expectedPause: true,
		},
		{
			name:          "valid pause with ratio 100",
			err:           fmt.Errorf("PAUSE:100"),
			expectedRatio: 100,
			expectedPause: true,
		},
		{
			name:          "valid pause with ratio 0",
			err:           fmt.Errorf("PAUSE:0"),
			expectedRatio: 0,
			expectedPause: true,
		},
		{
			name:          "invalid ratio above 100",
			err:           fmt.Errorf("PAUSE:150"),
			expectedRatio: 0,
			expectedPause: false,
		},
		{
			name:          "invalid negative ratio",
			err:           fmt.Errorf("PAUSE:-10"),
			expectedRatio: 0,
			expectedPause: false,
		},
		{
			name:          "invalid non-numeric ratio",
			err:           fmt.Errorf("PAUSE:abc"),
			expectedRatio: 0,
			expectedPause: false,
		},
		{
			name:          "resume command",
			err:           fmt.Errorf("RESUME"),
			expectedRatio: 0,
			expectedPause: false,
		},
		{
			name:          "other error",
			err:           fmt.Errorf("some other error"),
			expectedRatio: 0,
			expectedPause: false,
		},
		{
			name:          "empty pause command",
			err:           fmt.Errorf("PAUSE:"),
			expectedRatio: 0,
			expectedPause: false,
		},
		{
			name:          "pause without colon",
			err:           fmt.Errorf("PAUSE"),
			expectedRatio: 0,
			expectedPause: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio, pause := parseTrafficControlResponse(tt.err)
			assert.Equal(t, tt.expectedRatio, ratio)
			assert.Equal(t, tt.expectedPause, pause)
		})
	}
}

func TestSetManualTrafficControlState(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	ctrl := f.ctrl
	err := ctrl.setManualTrafficControlState(canary, f.deployer, 25)
	require.NoError(t, err)

	assert.Equal(t, flaggerv1.CanaryPhaseWaiting, canary.Status.Phase)
	assert.Equal(t, 25, canary.Status.CanaryWeight)
}

func TestSetManualTrafficControlState_AlreadyWaiting(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseWaiting

	ctrl := f.ctrl
	err := ctrl.setManualTrafficControlState(canary, f.deployer, 30)
	require.NoError(t, err)

	assert.Equal(t, flaggerv1.CanaryPhaseWaiting, canary.Status.Phase)
	assert.Equal(t, 30, canary.Status.CanaryWeight)
}

func TestClearManualTrafficControlState(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseWaiting

	ctrl := f.ctrl
	err := ctrl.clearManualTrafficControlState(canary, f.deployer)
	require.NoError(t, err)

	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canary.Status.Phase)
}

func TestClearManualTrafficControlState_NotWaiting(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	ctrl := f.ctrl
	err := ctrl.clearManualTrafficControlState(canary, f.deployer)
	require.NoError(t, err)

	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canary.Status.Phase)
}

func TestRunManualTrafficControlHooks_Pause(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("PAUSE:30"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue, ratio := f.ctrl.runManualTrafficControlHooks(canary, f.deployer, f.router)

	assert.False(t, shouldContinue)
	assert.Equal(t, 0, ratio)
	assert.Equal(t, flaggerv1.CanaryPhaseWaiting, canary.Status.Phase)
}

func TestRunManualTrafficControlHooks_Resume(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseWaiting

	shouldContinue, ratio := f.ctrl.runManualTrafficControlHooks(canary, f.deployer, f.router)

	assert.True(t, shouldContinue)
	assert.Equal(t, 0, ratio)
	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canary.Status.Phase)
}

func TestRunManualTrafficControlHooks_InvalidResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("PAUSE:invalid"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue, ratio := f.ctrl.runManualTrafficControlHooks(canary, f.deployer, f.router)

	assert.True(t, shouldContinue)
	assert.Equal(t, 0, ratio)
	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canary.Status.Phase)
}

func TestRunManualTrafficControlHooks_NoManualHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue, ratio := f.ctrl.runManualTrafficControlHooks(canary, f.deployer, f.router)

	assert.True(t, shouldContinue)
	assert.Equal(t, 0, ratio)
	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canary.Status.Phase)
}

func TestRunManualTrafficControlHooks_MultipleHooks(t *testing.T) {
	// First webhook returns success (resume), second returns pause
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("PAUSE:40"))
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

	// Since the first webhook succeeds, it should resume and return true
	shouldContinue, ratio := f.ctrl.runManualTrafficControlHooks(canaryObj, f.deployer, f.router)

	assert.True(t, shouldContinue)
	assert.Equal(t, 0, ratio)
	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canaryObj.Status.Phase)
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

	shouldContinue, ratio := f.ctrl.runManualTrafficControlHooks(canary, f.deployer, f.router)

	assert.True(t, shouldContinue)
	assert.Equal(t, 0, ratio)
	assert.Equal(t, flaggerv1.CanaryPhaseProgressing, canary.Status.Phase)
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
