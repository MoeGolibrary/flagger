#!/usr/bin/env bash

# Negative test for confirm-promotion webhook (testing timeout scenario)

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Negative Test: Confirm Promotion webhook timeout'

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
    iterations: 2
    metrics:
    - name: request-success-rate
      thresholdRange:
        min: 99
      interval: 1m
    webhooks:
      - name: confirm-promotion
        type: confirm-promotion
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
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.6"

wait_for_phase "Progressing"

echo '>>> Waiting for canary to reach waiting for promotion phase'
wait_for_phase "WaitingPromotion"

echo '>>> NOT approving promotion - testing timeout behavior'
echo '>>> Waiting for canary to fail due to webhook timeout'

# Wait for failure due to timeout
retries=30  # Longer wait time for timeout
count=0
ok=false
until ${ok}; do
    if kubectl -n test get canary/podinfo | grep 'Failed'; then
        echo '>>> Canary failed as expected due to webhook timeout'
        ok=true
    elif kubectl -n test get canary/podinfo | grep -v 'WaitingPromotion'; then
        # If it moved past WaitingPromotion, the test failed
        echo '>>> Canary moved past WaitingPromotion phase - test failed'
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

echo '✔ Confirm Promotion webhook timeout test passed'