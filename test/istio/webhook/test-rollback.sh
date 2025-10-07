#!/usr/bin/env bash

# Test for rollback webhook

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Rollback via webhook'

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
      - name: rollback-hook
        type: rollback
        url: http://flagger-loadtester.test/rollback/check
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 2m -q 10 -c 2 http://podinfo.test:9898/"
EOF
)

create_canary "$CANARY_SPEC"
wait_for_initialized
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.3"

wait_for_phase "Progressing"

echo '>>> Opening rollback gate to trigger rollback'
kubectl -n test exec deployment/flagger-loadtester -- curl -d '{"name": "podinfo","namespace":"test"}' http://localhost:8080/rollback/open

echo '>>> Waiting for canary to rollback'
wait_for_phase "Failed"
verify_canary_failure

echo '✔ Rollback via webhook test passed'