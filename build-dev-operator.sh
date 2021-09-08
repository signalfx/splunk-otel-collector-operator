#!/bin/bash

set -e
set -u

VERSION=0.33.${VERSION:-$(date +%s)}
BUNDLE_VERSION=${BUNDLE_VERSION:-$VERSION}
OPERATOR_VERSION=${BUNDLE_VERSION:-$OPERATOR_VERSION}

export IMG_PREFIX=quay.io/$QUAY_USERNAME
export USER=$QUAY_USERNAME
export IMG=quay.io/$USER/splunk-otel-operator:v$OPERATOR_VERSION
export BUNDLE_IMG=quay.io/$USER/splunk-otel-operator-bundle:v$BUNDLE_VERSION

build() {
    make generate
	make set-image-controller
	make container
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

publish() {
	docker push $IMG
	docker push $BUNDLE_IMG
	echo "==== Published the following images:"
	echo $BUNDLE_IMG
	echo $IMG 
}


deploy_local() {
	kubens operators
	make set-image-controller
	make container 
	make deploy
}


for arg; do
   "$arg"
done

