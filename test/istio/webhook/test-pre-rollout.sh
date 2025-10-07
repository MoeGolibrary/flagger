#!/usr/bin/env bash

# Test for pre-rollout webhook functionality

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Pre-rollout webhook'

initialize_test_workloads

CANARY_SPEC=$(cat <<EOF
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: podinfo
  namespace: test
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: podinfo
  progressDeadlineSeconds: 60
  service:
    port: 9898
    portDiscovery: true
  analysis:
    interval: 10s
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
      interval: 1m
    webhooks:
      - name: pre-rollout-test
        type: pre-rollout
        url: http://flagger-loadtester.test/
        timeout: 30s
        metadata:
          type: bash
          cmd: "echo 'Pre-rollout check passed'"
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 2m -q 10 -c 2 http://podinfo.test:9898/"
EOF
)

create_canary "$CANARY_SPEC"
wait_for_initialized
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.9"

echo '>>> Waiting for canary to progress'
wait_for_phase "Progressing"

# Wait a bit to make sure the pre-rollout webhook was executed
sleep 20

# Check if the canary is still progressing (which means pre-rollout passed)
current_phase=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.phase}')
if [[ "$current_phase" != "Progressing" && "$current_phase" != "WaitingPromotion" && "$current_phase" != "Succeeded" ]]; then
    echo ">>> Canary failed during pre-rollout phase: $current_phase"
    kubectl -n istio-system logs deployment/flagger --tail 20
    exit 1
fi

echo '>>> Waiting for canary completion'
wait_for_completion

echo '✔ Pre-rollout webhook test passed'