#!/bin/bash

set -e
set -u

VERSION=0.33.${VERSION:-$(date +%s)}
BUNDLE_VERSION=${BUNDLE_VERSION:-$VERSION}
OPERATOR_VERSION=${BUNDLE_VERSION:-$OPERATOR_VERSION}

export USER=${QUAY_USERNAME:-signalfx}
export IMG_PREFIX=quay.io/$USER
export IMG=quay.io/$USER/splunk-otel-operator:v$OPERATOR_VERSION
export BUNDLE_IMG=quay.io/$USER/splunk-otel-operator-bundle:v$BUNDLE_VERSION

build() {
#	make set-image-controller
	make generate
	make container
}

pack() {
	make bundle VERSION=${OPERATOR_VERSION}
	make bundle-build VERSION=${OPERATOR_VERSION}
	make release-artifacts
}


load() {
	kind load docker-image $IMG
	kind load docker-image $BUNDLE_IMG
}


install() {
    make cert-manager
    kubectl apply -f dist/splunk-otel-operator.yaml
}

build_install() {
	build
	pack
	load
	install
}

for arg; do
   "$arg"
done

