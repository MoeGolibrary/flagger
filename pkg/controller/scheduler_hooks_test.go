package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

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

// TestRunConfirmRolloutHooks tests the confirm-rollout webhook functionality
func TestRunConfirmRolloutHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmRolloutHooks(canary, f.deployer)

	assert.True(t, shouldContinue)
}

func TestRunConfirmRolloutHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "confirm-rollout",
			URL:  ts.URL,
			Type: flaggerv1.ConfirmRolloutHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmRolloutHooks(canaryObj, f.deployer)

	assert.True(t, shouldContinue)
}

func TestRunConfirmRolloutHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "confirm-rollout",
			URL:  ts.URL,
			Type: flaggerv1.ConfirmRolloutHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmRolloutHooks(canaryObj, f.deployer)

	assert.False(t, shouldContinue)
}

// TestRunConfirmPromotionHooks tests the confirm-promotion webhook functionality
func TestRunConfirmPromotionHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmPromotionHooks(canary, f.deployer)

	assert.True(t, shouldContinue)
}

func TestRunConfirmPromotionHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "confirm-promotion",
			URL:  ts.URL,
			Type: flaggerv1.ConfirmPromotionHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmPromotionHooks(canaryObj, f.deployer)

	assert.True(t, shouldContinue)
}

func TestRunConfirmPromotionHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "confirm-promotion",
			URL:  ts.URL,
			Type: flaggerv1.ConfirmPromotionHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runConfirmPromotionHooks(canaryObj, f.deployer)

	assert.False(t, shouldContinue)
}

// TestRunPreRolloutHooks tests the pre-rollout webhook functionality
func TestRunPreRolloutHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runPreRolloutHooks(canary)

	assert.True(t, shouldContinue)
}

func TestRunPreRolloutHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "pre-rollout",
			URL:  ts.URL,
			Type: flaggerv1.PreRolloutHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runPreRolloutHooks(canaryObj)

	assert.True(t, shouldContinue)
}

func TestRunPreRolloutHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "pre-rollout",
			URL:  ts.URL,
			Type: flaggerv1.PreRolloutHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldContinue := f.ctrl.runPreRolloutHooks(canaryObj)

	assert.False(t, shouldContinue)
}

// TestRunPostRolloutHooks tests the post-rollout webhook functionality
func TestRunPostRolloutHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseSucceeded

	shouldContinue := f.ctrl.runPostRolloutHooks(canary, flaggerv1.CanaryPhaseSucceeded)

	assert.True(t, shouldContinue)
}

func TestRunPostRolloutHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "post-rollout",
			URL:  ts.URL,
			Type: flaggerv1.PostRolloutHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseSucceeded

	shouldContinue := f.ctrl.runPostRolloutHooks(canaryObj, flaggerv1.CanaryPhaseSucceeded)

	assert.True(t, shouldContinue)
}

func TestRunPostRolloutHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "post-rollout",
			URL:  ts.URL,
			Type: flaggerv1.PostRolloutHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseSucceeded

	shouldContinue := f.ctrl.runPostRolloutHooks(canaryObj, flaggerv1.CanaryPhaseSucceeded)

	assert.False(t, shouldContinue)
}

// TestRunRollbackHooks tests the rollback webhook functionality
func TestRunRollbackHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldRollback := f.ctrl.runRollbackHooks(canary, flaggerv1.CanaryPhaseProgressing)

	assert.False(t, shouldRollback)
}

func TestRunRollbackHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "rollback",
			URL:  ts.URL,
			Type: flaggerv1.RollbackHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldRollback := f.ctrl.runRollbackHooks(canaryObj, flaggerv1.CanaryPhaseProgressing)

	assert.True(t, shouldRollback)
}

func TestRunRollbackHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "rollback",
			URL:  ts.URL,
			Type: flaggerv1.RollbackHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldRollback := f.ctrl.runRollbackHooks(canaryObj, flaggerv1.CanaryPhaseProgressing)

	assert.False(t, shouldRollback)
}

// TestRunSkipHooks tests the skip webhook functionality
func TestRunSkipHooks_NoHooks(t *testing.T) {
	f := newDeploymentFixture(nil)

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldSkip := f.ctrl.runSkipHooks(canary, flaggerv1.CanaryPhaseProgressing)

	assert.False(t, shouldSkip)
}

func TestRunSkipHooks_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "skip",
			URL:  ts.URL,
			Type: flaggerv1.SkipHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldSkip := f.ctrl.runSkipHooks(canaryObj, flaggerv1.CanaryPhaseProgressing)

	assert.True(t, shouldSkip)
}

func TestRunSkipHooks_Failure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	canary := newDeploymentTestCanary()
	canary.Spec.Analysis.Webhooks = []flaggerv1.CanaryWebhook{
		{
			Name: "skip",
			URL:  ts.URL,
			Type: flaggerv1.SkipHook,
		},
	}

	f := newDeploymentFixture(canary)

	canaryObj := f.canary.DeepCopy()
	canaryObj.Status.Phase = flaggerv1.CanaryPhaseProgressing

	shouldSkip := f.ctrl.runSkipHooks(canaryObj, flaggerv1.CanaryPhaseProgressing)

	assert.False(t, shouldSkip)
}

// TestRunManualTrafficControlHooks_WeightOnly tests manual traffic control with only weight specified
func TestRunManualTrafficControlHooks_WeightOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"weight\": 50, \"timestamp\": \"1234567890\"}"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.NoError(t, err)

	assert.NotNil(t, manualState)
	assert.Equal(t, 50, *manualState.Weight)
	// Paused should be false by default when not specified
	assert.False(t, manualState.Paused)
}

// TestRunManualTrafficControlHooks_PausedOnly tests manual traffic control with only paused specified
func TestRunManualTrafficControlHooks_PausedOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"paused\": true, \"timestamp\": \"1234567890\"}"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.NoError(t, err)

	assert.NotNil(t, manualState)
	// Weight should be nil when not specified
	assert.Nil(t, manualState.Weight)
	assert.True(t, manualState.Paused)
}

// TestRunManualTrafficControlHooks_EmptyResponse tests manual traffic control with empty response
func TestRunManualTrafficControlHooks_EmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer ts.Close()

	f := newDeploymentFixture(newDeploymentTestCanaryWithManualHook(ts.URL))

	canary := f.canary.DeepCopy()
	canary.Status.Phase = flaggerv1.CanaryPhaseProgressing

	manualState, err := f.ctrl.runManualTrafficControlHooks(canary)
	require.NoError(t, err)

	assert.NotNil(t, manualState)
	// Both weight and paused should be their zero values when not specified
	assert.Nil(t, manualState.Weight)
	assert.False(t, manualState.Paused)
}

// TestRunManualTrafficControlHooks_NoTimestamp tests manual traffic control without timestamp
func TestRunManualTrafficControlHooks_NoTimestamp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"weight\": 30, \"paused\": true}"))
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
	assert.Equal(t, "", manualState.Timestamp)
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
