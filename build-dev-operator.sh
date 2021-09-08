#!/bin/bash

set -e
set -u

VERSION_FROM_FILE="$(grep -v '\#' versions.txt | grep operator | awk -F= '{print $2}')"

export VERSION=${VERSION:-$VERSION_FROM_FILE}
export USER=${QUAY_USERNAME:-signalfx}
export IMG_PREFIX=quay.io/$USER
export IMG=quay.io/$USER/splunk-otel-operator:v$VERSION

BUNDLE_VERSION=${BUNDLE_VERSION:-$VERSION}
export BUNDLE_IMG=quay.io/$USER/splunk-otel-operator-bundle:v$BUNDLE_VERSION

build() {
#	make set-image-controller
	make generate
	make container
}

pack() {
	make bundle
	make bundle-build
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

