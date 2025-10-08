#!/usr/bin/env bash

# Test for manual traffic control webhook

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Manual Traffic Control via webhook'

# Initialize test workloads
initialize_test_workloads

# Create canary with manual traffic control webhook
echo '>>> Creating canary with manual traffic control webhook'
kubectl apply -f - <<EOF
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
  service:
    port: 9898
    gateways:
    - istio-system/public-gateway
    hosts:
    - app.example.com
  analysis:
    interval: 10s
    threshold: 5
    maxWeight: 100
    stepWeight: 20
    webhooks:
    - name: manual-traffic-control
      type: manual-traffic-control
      url: http://flagger-loadtester.test/traffic/
EOF

# Wait for canary to be initialized
wait_for_initialized

# Trigger a deployment to start canary analysis
trigger_deployment stefanprodan/podinfo:3.1.1

# Wait for canary to start progressing
wait_for_phase Progressing

# Send command to pause at 40% weight
echo '>>> Sending manual traffic control command to pause at 40% weight'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"weight": 40, "paused": true}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

echo '>>> Resuming canary'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"paused": false}' http://podinfo-canary:9898/traffic/

# Wait for canary to complete
wait_for_completion

echo '✔ Manual Traffic Control via webhook test passed'