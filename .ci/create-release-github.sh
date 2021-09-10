#!/bin/bash

OPERATOR_VERSION=$(git describe --tags)
# echo "${GH_WRITE_TOKEN}" | gh auth login --with-token

# gh config set prompt disabled
gh release create \
    -t "Release ${OPERATOR_VERSION}" \
    "${OPERATOR_VERSION}" \
    'dist/splunk-otel-operator.yaml#Installation manifest for Kubernetes' \
    'dist/splunk-otel-operator-openshift.yaml#Installation manifest for OpenShift'
