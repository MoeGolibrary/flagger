#!/usr/bin/env bash

# Test for manual traffic control webhook with multiple resume operations
# This test validates that when resuming from a paused state multiple times,
# the canary weight is maintained correctly throughout all operations.

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Manual Traffic Control Multi Resume Operations'

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
    stepWeight: 10
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

# Send command to set weight to 30% and pause
echo '>>> Sending manual traffic control command to set weight to 30% and pause'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"weight": 30, "paused": true}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

# Check that the canary is paused at 30% weight
echo '>>> Checking that canary is paused at 30% weight'
kubectl -n test get canary/podinfo | grep 'Waiting'
weight=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.canaryWeight}')
if [ "$weight" != "30" ]; then
  echo "ERROR: Expected weight to be 30, but got $weight"
  exit 1
fi
echo "✓ Canary is paused at 30% weight"

# Send command to resume (without specifying weight)
echo '>>> Sending manual traffic control command to resume without specifying weight'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"paused": false}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

# Check that the canary is progressing and still at 30% weight
echo '>>> Checking that canary is progressing and maintains 30% weight'
kubectl -n test get canary/podinfo | grep 'Progressing'
weight=$(kubectl -n test get canary/podinfo -o jsonpath='{.status.canaryWeight}')
if [ "$weight" != "30" ]; then
  echo "ERROR: Expected weight to remain at 30 after resume, but got $weight"
  exit 1
fi
echo "✓ Canary is progressing and maintains 30% weight after first resume"

# Let it progress a bit
sleep 20

# Pause again at current weight
echo '>>> Pausing at current weight'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"paused": true}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

# Check that the canary is paused
echo '>>> Checking that canary is paused'
kubectl -n test get canary/podinfo | grep 'Waiting'
echo "✓ Canary is paused at current weight"

# Resume again
echo '>>> Resuming again'
kubectl -n test exec deployment/flagger-loadtester -- curl -s -d '{"paused": false}' http://podinfo-canary:9898/traffic/

# Wait a bit for the command to be processed
sleep 15

# Check that the canary is progressing
echo '>>> Checking that canary is progressing'
kubectl -n test get canary/podinfo | grep 'Progressing'
echo "✓ Canary is progressing after second resume"

# Wait for canary to complete
wait_for_completion

echo '✔ Manual Traffic Control Multi Resume Operations test passed'