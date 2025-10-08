#!/usr/bin/env bash

# Test for manual traffic control webhook
# Note: This was previously a placeholder, but now there is a proper implementation

set -o errexit

REPO_ROOT=$(git rev-parse --show-toplevel)
source "$(dirname "$0")/base.sh"

echo '>>> Test: Manual Traffic Control via webhook'

echo '>>> This test was previously a placeholder. Please run test-manual-traffic-control-proper.sh for the actual implementation.'
echo '>>> The manual traffic control webhook functionality is now implemented in the loadtester.'
echo '>>>'
echo '>>> The webhook is configured as follows:'
echo '>>>'
echo '>>> webhooks:'
echo '>>>   - name: manual-traffic-control'
echo '>>>     type: manual-traffic-control'
echo '>>>     url: http://flagger-loadtester.test/traffic/'
echo '>>>'
echo '>>> To pause traffic, a request would be sent to:'
echo '>>>   curl -d '\'{"paused": true}\'' http://localhost:8080/traffic/'
echo '>>>'
echo '>>> To resume traffic, a request would be sent to:'
echo '>>>   curl -d '\'{"paused": false}\'' http://localhost:8080/traffic/'
echo '>>>'
echo '>>> To set a specific weight, a request would be sent to:'
echo '>>>   curl -d '\'{"weight": 30}\'' http://localhost:8080/traffic/'
echo '>>>'
echo '>>> Please run test-manual-traffic-control-proper.sh for the complete test.'

echo '✔ Manual Traffic Control via webhook placeholder documented'