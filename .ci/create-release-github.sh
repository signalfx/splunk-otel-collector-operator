#!/bin/bash

OPERATOR_VERSION=$(git describe --tags)

# gh config set prompt disabled
gh release create \
    -R https://github.com/signalfx/splunk-otel-collector-operator \
    -t "Release ${OPERATOR_VERSION}" \
    "${OPERATOR_VERSION}" \
    'dist/splunk-otel-operator.yaml#Installation manifest for Kubernetes' \
    'dist/splunk-otel-operator-openshift.yaml#Installation manifest for OpenShift'
