# Istio Webhook E2E Tests

This directory contains end-to-end tests for Flagger's webhook functionality when used with Istio.

## Overview

Flagger supports various webhook types that can be used to extend the canary analysis and control the deployment process. These tests validate that the webhooks work correctly in an Istio environment.

## Webhook Types

### Implemented Webhooks

1. **confirm-rollout** - Pauses the rollout until approved
2. **confirm-promotion** - Pauses the promotion until approved
3. **rollback** - Triggers rollback when called
4. **manual-traffic-control** - Allows manual control of traffic routing

### Placeholder Webhooks

The following webhook types are defined in the Flagger API but not yet implemented in the loadtester:

1. **skip** - Would skip the analysis phase

## Test Organization

Each webhook type has its own test script:

- `test-confirm-rollout.sh` - Tests the confirm-rollout webhook
- `test-confirm-promotion.sh` - Tests the confirm-promotion webhook
- `test-rollback.sh` - Tests the rollback webhook
- `test-confirm-rollout-failure.sh` - Tests confirm-rollout webhook timeout behavior
- `test-invalid-webhook.sh` - Tests behavior with invalid webhook configurations
- `test-skip.sh` - Placeholder for skip webhook test
- `test-manual-traffic-control.sh` - Tests basic manual traffic control webhook
- `test-manual-traffic-control-proper.sh` - Tests manual traffic control with proper weight setting
- `test-manual-traffic-control-resume.sh` - Tests manual traffic control resume with weight maintenance
- `test-manual-traffic-control-multi-resume.sh` - Tests manual traffic control with multiple resume operations

## Running the Tests

To run all webhook tests:

```bash
cd test/istio/webhook
./run.sh
```

To run a specific test:

```bash
cd test/istio/webhook
./test-confirm-rollout.sh
```

## Implementation Details

### Confirm Rollout Webhook

This webhook pauses the canary rollout until approved:

```yaml
webhooks:
  - name: confirm-rollout
    type: confirm-rollout
    url: http://flagger-loadtester.test/gate/check
```

To approve the rollout:
```bash
curl -d '{"name": "podinfo","namespace":"test"}' http://localhost:8080/gate/open
```

### Confirm Promotion Webhook

This webhook pauses the canary promotion until approved:

```yaml
webhooks:
  - name: confirm-promotion
    type: confirm-promotion
    url: http://flagger-loadtester.test/gate/check
```

To approve the promotion:
```bash
curl -d '{"name": "podinfo","namespace":"test"}' http://localhost:8080/gate/approve
```

### Rollback Webhook

This webhook triggers a rollback when called:

```yaml
webhooks:
  - name: rollback-hook
    type: rollback
    url: http://flagger-loadtester.test/rollback/check
```

To trigger a rollback:
```bash
curl -d '{"name": "podinfo","namespace":"test"}' http://localhost:8080/rollback/open
```

### Manual Traffic Control Webhook

This webhook allows manual control of traffic routing during canary analysis:

```yaml
webhooks:
  - name: manual-traffic-control
    type: manual-traffic-control
    url: http://flagger-loadtester.test/traffic/
```

To set traffic weight and pause:
```bash
curl -d '{"weight": 40, "paused": true}' http://podinfo-canary:9898/traffic/
```

To resume traffic shifting:
```bash
curl -d '{"paused": false}' http://podinfo-canary:9898/traffic/
```

## Failure Scenarios

The tests also include negative test cases to verify behavior in failure scenarios:

1. **Webhook timeout** - Tests what happens when a webhook doesn't respond within the timeout period
2. **Invalid webhook configuration** - Tests behavior when webhooks are misconfigured

## Future Enhancements

To fully test all webhook scenarios, the loadtester would need to implement additional endpoints as documented in the placeholder test scripts.