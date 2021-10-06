#!/bin/bash

os=$(go env GOOS)
arch=$(go env GOARCH)
curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.1.0/kubebuilder_3.1.0_${os}_${arch}.tar.gz | tar -xz -C /tmp/
sudo mv /tmp/kubebuilder_3.1.0_${os}_${arch} /usr/local/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin
