#!/usr/bin/env bash
set -e

SCRIPT_NAME=arena
SCRIPT_FILES=/etc/arena
SCRIPT_CHARTS=${SCRIPT_FILES}/charts

SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"

cp -rf "$SCRIPT_DIR/bin/arena" /usr/local/bin/${SCRIPT_NAME}

if [ -d "${SCRIPT_FILES}" ]; then
rm -rf ${SCRIPT_FILES}
fi

mkdir ${SCRIPT_FILES}
cp -R "${SCRIPT_DIR}/charts" ${SCRIPT_CHARTS}