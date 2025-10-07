#!/usr/bin/env bash

# Negative test for confirm-rollout webhook (testing timeout scenario)

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Negative Test: Confirm Rollout webhook timeout'

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
        timeout: 15s  # Short timeout to test timeout behavior
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 2m -q 10 -c 2 http://podinfo.test:9898/"
EOF
)

create_canary "$CANARY_SPEC"
wait_for_initialized
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.4"

echo '>>> Waiting for canary to wait for rollout confirmation'
wait_for_phase "Waiting"

echo '>>> NOT approving rollout - testing timeout behavior'
echo '>>> Waiting for canary to fail due to webhook timeout'

# Wait for failure due to timeout
retries=30  # Longer wait time for timeout
count=0
ok=false
until ${ok}; do
    if kubectl -n test get canary/podinfo | grep 'Failed'; then
        echo '>>> Canary failed as expected due to webhook timeout'
        ok=true
    elif kubectl -n test get canary/podinfo | grep -v 'Waiting'; then
        # If it moved past waiting, the test failed
        echo '>>> Canary moved past waiting phase - test failed'
        kubectl -n test get canary/podinfo
        exit 1
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        kubectl -n istio-system logs deployment/flagger
        echo "No more retries left - test failed"
        exit 1
    fi
done

verify_canary_failure

echo '✔ Confirm Rollout webhook timeout test passed'