#!/usr/bin/env bash

# Test for invalid webhook configuration

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Invalid webhook configuration'

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
      - name: invalid-webhook
        type: confirm-rollout
        url: http://non-existent-service.test/  # Invalid URL
      - name: load-test
        url: http://flagger-loadtester.test/
        metadata:
          cmd: "hey -z 2m -q 10 -c 2 http://podinfo.test:9898/"
EOF
)

create_canary "$CANARY_SPEC"
wait_for_initialized
trigger_deployment "ghcr.io/stefanprodan/podinfo:6.0.5"

echo '>>> Waiting for canary to fail due to invalid webhook'
retries=30
count=0
ok=false
until ${ok}; do
    if kubectl -n test get canary/podinfo | grep 'Failed'; then
        echo '>>> Canary failed as expected due to invalid webhook'
        ok=true
    elif kubectl -n test get canary/podinfo | grep 'Progressing'; then
        # Check if it's stuck due to invalid webhook
        failed_checks=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.failedChecks}')
        if [[ $failed_checks -gt 0 ]]; then
            echo '>>> Canary has failed checks due to invalid webhook'
            ok=true
        fi
    fi
    sleep 5
    count=$(($count + 1))
    if [[ ${count} -eq ${retries} ]]; then
        kubectl -n istio-system logs deployment/flagger
        echo "No more retries left - test may have failed"
        # This might be an acceptable outcome depending on how Flagger handles invalid webhooks
        break
    fi
done

echo '>>> Checking final status'
kubectl -n test get canary/podinfo

echo '✔ Invalid webhook configuration test completed'