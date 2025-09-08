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
	"strconv"
	"strings"

	"go.uber.org/zap/zapcore"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/fluxcd/flagger/pkg/canary"
	"github.com/fluxcd/flagger/pkg/router"
)

func (c *Controller) runConfirmTrafficIncreaseHooks(canary *flaggerv1.Canary) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.ConfirmTrafficIncreaseHook {
			err := CallWebhook(*canary, flaggerv1.CanaryPhaseProgressing, webhook)
			if err != nil {
				c.recordEventWarningf(canary, "Halt %s.%s advancement waiting for traffic increase approval %s",
					canary.Name, canary.Namespace, webhook.Name)
				if !webhook.MuteAlert {
					c.alert(canary, fmt.Sprintf("Halt %s.%s advancement waiting for traffic increase approval %s", canary.Name, canary.Namespace, webhook.Name), false, flaggerv1.SeverityWarn)
				}
				return false
			}
			c.recordEventInfof(canary, "Confirm-traffic-increase check %s passed", webhook.Name)
		}
	}
	return true
}

func (c *Controller) runConfirmRolloutHooks(canary *flaggerv1.Canary, canaryController canary.Controller) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.ConfirmRolloutHook {
			err := CallWebhook(*canary, canary.Status.Phase, webhook)
			if err != nil {
				if canary.Status.Phase != flaggerv1.CanaryPhaseWaiting {
					if err := canaryController.SetStatusPhase(canary, flaggerv1.CanaryPhaseWaiting); err != nil {
						c.logCanaryEvent(canary, fmt.Sprintf("%v", err), zapcore.ErrorLevel)
					}
					c.recordEventWarningf(canary, "Halt %s.%s advancement waiting for approval %s",
						canary.Name, canary.Namespace, webhook.Name)
					if !webhook.MuteAlert {
						c.alert(canary, fmt.Sprintf("Halt %s.%s advancement waiting for approval %s",
							canary.Name, canary.Namespace, webhook.Name), false, flaggerv1.SeverityWarn)
					}
				}
				return false
			}
			c.recordEventInfof(canary, "Confirm-rollout check %s passed", webhook.Name)
		}
	}
	return true
}

func (c *Controller) runConfirmPromotionHooks(canary *flaggerv1.Canary, canaryController canary.Controller) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.ConfirmPromotionHook {
			err := CallWebhook(*canary, flaggerv1.CanaryPhaseProgressing, webhook)
			if err != nil {
				if canary.Status.Phase != flaggerv1.CanaryPhaseWaitingPromotion {
					if err := canaryController.SetStatusPhase(canary, flaggerv1.CanaryPhaseWaitingPromotion); err != nil {
						c.logCanaryEvent(canary, fmt.Sprintf("%v", err), zapcore.ErrorLevel)
					}
					c.recordEventWarningf(canary, "Halt %s.%s advancement waiting for promotion approval %s",
						canary.Name, canary.Namespace, webhook.Name)
					if !webhook.MuteAlert {
						c.alert(canary, fmt.Sprintf("Halt %s.%s advancement waiting for promotion approval %s",
							canary.Name, canary.Namespace, webhook.Name), false, flaggerv1.SeverityWarn)
					}
				} else {
					if err := canaryController.SetStatusIterations(canary, canary.GetAnalysis().Iterations-1); err != nil {
						c.recordEventWarningf(canary, "%v", err)
					}
				}
				return false
			} else {
				c.recordEventInfof(canary, "Confirm-promotion check %s passed", webhook.Name)
			}
		}
	}
	return true
}

func (c *Controller) runPreRolloutHooks(canary *flaggerv1.Canary) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.PreRolloutHook {
			err := CallWebhook(*canary, flaggerv1.CanaryPhaseProgressing, webhook)
			if err != nil {
				c.recordEventWarningf(canary, "Halt %s.%s advancement pre-rollout check %s failed %v",
					canary.Name, canary.Namespace, webhook.Name, err)
				return false
			} else {
				c.recordEventInfof(canary, "Pre-rollout check %s passed", webhook.Name)
			}
		}
	}
	return true
}

func (c *Controller) runPostRolloutHooks(canary *flaggerv1.Canary, phase flaggerv1.CanaryPhase) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.PostRolloutHook {
			err := CallWebhook(*canary, phase, webhook)
			if err != nil {
				c.recordEventWarningf(canary, "Post-rollout hook %s failed %v", webhook.Name, err)
				return false
			} else {
				c.recordEventInfof(canary, "Post-rollout check %s passed", webhook.Name)
			}
		}
	}
	return true
}

func (c *Controller) runRollbackHooks(canary *flaggerv1.Canary, phase flaggerv1.CanaryPhase) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.RollbackHook {
			err := CallWebhook(*canary, phase, webhook)
			if err != nil {
				c.recordEventInfof(canary, "Rollback hook %s not signaling a rollback", webhook.Name)
			} else {
				c.recordEventWarningf(canary, "Rollback check %s passed", webhook.Name)
				return true
			}
		}
	}
	return false
}

func (c *Controller) runSkipHooks(canary *flaggerv1.Canary, phase flaggerv1.CanaryPhase) bool {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.SkipHook {
			err := CallWebhook(*canary, phase, webhook)
			if err != nil {
				c.recordEventInfof(canary, "Skip Canary hook %s not signaling a rollback", webhook.Name)
			} else {
				c.recordEventWarningf(canary, "Skip Canary check %s passed", webhook.Name)
				return true
			}
		}
	}
	return false
}

func (c *Controller) runManualTrafficControlHooks(canary *flaggerv1.Canary, canaryController canary.Controller, meshRouter router.Interface) (shouldContinue bool, manualTrafficRatio int) {
	for _, webhook := range canary.GetAnalysis().Webhooks {
		if webhook.Type == flaggerv1.ManualTrafficControlHook {
			err := CallWebhook(*canary, canary.Status.Phase, webhook)
			if err != nil {
				if trafficRatio, shouldPause := parseTrafficControlResponse(err); shouldPause {
					if err := c.setManualTrafficControlState(canary, canaryController, trafficRatio); err != nil {
						c.recordEventWarningf(canary, "Failed to set manual traffic control: %v", err)
						return false, 0
					}

					primaryWeight := 100 - trafficRatio
					if err := meshRouter.SetRoutes(canary, primaryWeight, trafficRatio, false); err != nil {
						c.recordEventWarningf(canary, "Failed to set traffic routes: %v", err)
						return false, 0
					}

					c.recordEventInfof(canary, "Manual traffic control activated: %d%% canary traffic", trafficRatio)
					return false, trafficRatio
				}
			} else {
				if err := c.clearManualTrafficControlState(canary, canaryController); err != nil {
					c.recordEventWarningf(canary, "Failed to clear manual traffic control: %v", err)
				}
				c.recordEventInfof(canary, "Manual traffic control deactivated, resuming automatic progression")
				return true, 0
			}
		}
	}
	return true, 0
}

func parseTrafficControlResponse(err error) (int, bool) {
	errMsg := err.Error()
	if strings.HasPrefix(errMsg, "PAUSE:") {
		if ratio, parseErr := strconv.Atoi(strings.TrimPrefix(errMsg, "PAUSE:")); parseErr == nil {
			if ratio >= 0 && ratio <= 100 {
				return ratio, true
			}
		}
	}
	return 0, false
}

func (c *Controller) setManualTrafficControlState(canary *flaggerv1.Canary, canaryController canary.Controller, trafficRatio int) error {
	if canary.Status.Phase != flaggerv1.CanaryPhaseWaiting {
		if err := canaryController.SetStatusPhase(canary, flaggerv1.CanaryPhaseWaiting); err != nil {
			return err
		}
		c.recordEventInfof(canary, "Canary paused for manual traffic control")
	}

	return canaryController.SetStatusWeight(canary, trafficRatio)
}

func (c *Controller) clearManualTrafficControlState(canary *flaggerv1.Canary, canaryController canary.Controller) error {
	if canary.Status.Phase == flaggerv1.CanaryPhaseWaiting {
		if err := canaryController.SetStatusPhase(canary, flaggerv1.CanaryPhaseProgressing); err != nil {
			return err
		}
		c.recordEventInfof(canary, "Canary resumed from manual traffic control")
	}
	return nil
}
