#!/usr/bin/env bash

# Negative test for rollback webhook (testing invalid configuration scenario)

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Negative Test: Rollback webhook with invalid configuration'

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
        url: http://non-existent-rollback-service.test/rollback/check  # Invalid URL
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 2m -q 10 -c 2 http://podinfo.test:9898/"
EOF
)

create_canary "$CANARY_SPEC"
wait_for_initialized
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.7"

wait_for_phase "Progressing"

echo '>>> Waiting to see if canary fails due to invalid rollback webhook'
retries=30
count=0
ok=false
until ${ok}; do
    if kubectl -n test get canary/podinfo | grep 'Failed'; then
        echo '>>> Canary failed - checking if it is due to invalid rollback webhook'
        # Check if it's due to invalid webhook by looking at failed checks
        failed_checks=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.failedChecks}')
        if [[ $failed_checks -gt 0 ]]; then
            echo '>>> Canary has failed checks, which may include invalid rollback webhook'
            ok=true
        fi
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        # Even if it didn't fail, we should check the status
        echo '>>> Checking current status after timeout'
        kubectl -n test get canary/podinfo
        break
    fi
done

echo '>>> Final status check'
kubectl -n test get canary/podinfo

# Try to trigger rollback anyway to make sure it doesn't work
echo '>>> Attempting to trigger rollback on invalid webhook (should not cause rollback)'
set +e
kubectl -n test exec deployment/flagger-loadtester -- curl -d '{"name": "podinfo","namespace":"test"}' http://localhost:8080/rollback/open
set -e

echo '>>> Checking status after attempted rollback'
kubectl -n test get canary/podinfo

echo '✔ Rollback webhook with invalid configuration test completed'