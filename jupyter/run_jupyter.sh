#!/usr/bin/env bash
set -xe

CONF_DIR="$HOME/.jupyter"
mkdir -p $CONF_DIR

cp /jupyter_notebook_config.py $CONF_DIR

set +e
GIT_REPO=ai-starter
if [ ! -d "$GIT_REPO" ]; then
  git clone https://github.com/AliyunContainerService/$GIT_REPO.git
fi
set -e

jupyter notebook --allow-root "$@"