#!/bin/bash

set -e
set -u

VERSION=0.33.${VERSION:-$(date +%s)}
BUNDLE_VERSION=${BUNDLE_VERSION:-$VERSION}
OPERATOR_VERSION=${BUNDLE_VERSION:-$OPERATOR_VERSION}

export USER=${QUAY_USERNAME:-signalfx}
export IMG=quay.io/$USER/splunk-otel-operator:v$OPERATOR_VERSION
export BUNDLE_IMG=quay.io/$USER/splunk-otel-operator-bundle:v$BUNDLE_VERSION

build() {
    make generate
	make set-image-controller
	make container
	make bundle VERSION=${OPERATOR_VERSION}
	make bundle-build VERSION=${OPERATOR_VERSION}
}

build_bundle() {
	make bundle VERSION=${OPERATOR_VERSION}
}

publish() {
	kind load docker-image $IMG
	kind load docker-image $BUNDLE_IMG

	docker push $IMG
	docker push $BUNDLE_IMG

	echo "==== Published the following images:"
	echo $BUNDLE_IMG
	echo $IMG 
}

build_publish() {
	build
	publish
}

deploy_local() {
	kubens operators
	make set-image-controller
	make container 
	make deploy
}

"$@"
