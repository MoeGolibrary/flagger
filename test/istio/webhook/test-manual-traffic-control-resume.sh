#!/usr/bin/env bash

# Test for manual traffic control webhook with resume functionality
# This test specifically validates that when resuming from a paused state,
# the canary weight is maintained rather than reset to 0.

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Manual Traffic Control Resume with Weight Maintenance'

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

# Send command to set weight to 60% and pause
echo '>>> Sending manual traffic control command to set weight to 60% and pause'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"weight": 60, "paused": true}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

# Check that the canary is paused at 60% weight
echo '>>> Checking that canary is paused at 60% weight'
kubectl -n test get canary/podinfo | grep 'Waiting'
weight=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.canaryWeight}')
if [ "$weight" != "60" ]; then
  echo "ERROR: Expected weight to be 60, but got $weight"
  exit 1
fi
echo "✓ Canary is paused at 60% weight"

# Send command to resume (without specifying weight)
echo '>>> Sending manual traffic control command to resume without specifying weight'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"paused": false}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

# Check that the canary is progressing and still at 60% weight
echo '>>> Checking that canary is progressing and maintains 60% weight'
kubectl -n test get canary/podinfo | grep 'Progressing'
weight=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.canaryWeight}')
if [ "$weight" != "60" ]; then
  echo "ERROR: Expected weight to remain at 60 after resume, but got $weight"
  exit 1
fi
echo "✓ Canary is progressing and maintains 60% weight after resume"

# Wait for canary to complete
wait_for_completion

echo '✔ Manual Traffic Control Resume with Weight Maintenance test passed'