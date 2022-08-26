#!/bin/bash

os=$(go env GOOS)
arch=$(go env GOARCH)
curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.6.0/kubebuilder_${os}_${arch} > /tmp/kubebuilder
sudo mv /tmp/kubebuilder /usr/local
export PATH=$PATH:/usr/local/kubebuilder/bin
