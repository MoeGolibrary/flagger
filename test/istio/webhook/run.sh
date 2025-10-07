#!/usr/bin/env bash

# Run all webhook E2E tests

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
DIR="$(cd "$(dirname "$0")" && pwd)"

echo ">>> Starting Istio Webhook E2E tests"

echo ">>> Running confirm-rollout webhook test"
"$DIR"/test-confirm-rollout.sh

echo ">>> Running confirm-promotion webhook test"
"$DIR"/test-confirm-promotion.sh

echo ">>> Running rollback webhook test"
"$DIR"/test-rollback.sh

echo ">>> Running confirm-rollout failure test (webhook timeout)"
"$DIR"/test-confirm-rollout-failure.sh

echo ">>> Running confirm-promotion failure test (webhook timeout)"
"$DIR"/test-confirm-promotion-failure.sh

echo ">>> Running invalid webhook test"
"$DIR"/test-invalid-webhook.sh

echo ">>> Running rollback webhook failure test"
"$DIR"/test-rollback-failure.sh

echo ">>> Running pre-rollout webhook test"
"$DIR"/test-pre-rollout.sh

echo ">>> Showing placeholder for skip webhook test"
"$DIR"/test-skip.sh

echo ">>> Showing placeholder for manual traffic control webhook test"
"$DIR"/test-manual-traffic-control.sh

echo ">>> All Istio Webhook E2E tests completed"