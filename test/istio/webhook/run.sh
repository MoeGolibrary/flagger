#!/usr/bin/env bash

# Script to run all webhook tests

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)

echo "Running webhook tests"

# Run webhook tests
tests=(
  "test-confirm-promotion.sh"
  "test-confirm-promotion-failure.sh"
  "test-confirm-rollout.sh"
  "test-confirm-rollout-failure.sh"
  "test-invalid-webhook.sh"
  "test-manual-traffic-control.sh"
  "test-manual-traffic-control-proper.sh"
  "test-manual-traffic-control-resume.sh"
  "test-manual-traffic-control-multi-resume.sh"
  "test-pre-rollout.sh"
  "test-rollback.sh"
  "test-rollback-failure.sh"
  "test-skip.sh"
)

for test in "${tests[@]}"; do
  echo "Running $test"
  "$REPO_ROOT/test/istio/webhook/$test"
done

echo "All webhook tests passed"