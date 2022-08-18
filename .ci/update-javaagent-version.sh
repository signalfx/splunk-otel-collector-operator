#!/usr/bin/env bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="${SCRIPT_DIR}/../"
cd ${ROOT_DIR}

print_usage() {
  cat <<EOF
Usage: $(basename $0) new_splunk_javaagent_version
All versions MUST NOT begin with 'v'. Example: 1.2.3".
EOF
}

if [[ $# < 1 ]]
then
  print_usage
  exit 1
fi

new_splunk_javaagent_version=$1

# MacOS requires passing backup file extension to in-place sed
if [[ $(uname -s) == "Darwin" ]]
then
  sed_flag='-i.tmp'
else
  sed_flag='-i'
fi

sed ${sed_flag} \
  -e "s/defaultJavaAgentVersion = \".*\"/defaultJavaAgentVersion = \"v$new_splunk_javaagent_version\"/" \
  apis/otel/v1alpha1/defaults.go