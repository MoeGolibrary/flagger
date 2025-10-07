#!/usr/bin/env bash

# Test for confirm-rollout webhook

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Confirm Rollout via webhook'

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
      - name: confirm-rollout
        type: confirm-rollout
        url: http://flagger-loadtester.test/gate/check
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 2m -q 10 -c 2 http://podinfo.test:9898/"
EOF
)

create_canary "$CANARY_SPEC"
wait_for_initialized
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.1"
wait_for_phase "Waiting"

echo '>>> Approving rollout via webhook'
kubectl -n test exec deployment/flagger-loadtester -- curl -d '{"name": "podinfo","namespace":"test"}' http://localhost:8080/gate/open

wait_for_phase "Progressing"
wait_for_completion

echo '✔ Confirm Rollout via webhook test passed'