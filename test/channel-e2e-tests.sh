#!/usr/bin/env bash

source $(dirname $0)/e2e-common.sh

go_test_e2e -tags=e2e -timeout=1h -run=TestKafkaChannelFirstEventDelay ./test/e2e_new_channel/... || fail_test "First event delay test failed"

go_test_e2e -timeout=1h ./test/e2e_channel/... -channels=messaging.knative.dev/v1beta1:KafkaChannel || fail_test "E2E suite (KafkaChannel) failed"
