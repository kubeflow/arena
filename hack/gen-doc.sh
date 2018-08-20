#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

PREFIX="arena"
CMD_PREFIX="cmd"

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

cd ${SCRIPT_ROOT}
echo "Generating CLI documentation..."
go run hack/docgen.go
cd - > /dev/null