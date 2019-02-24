#!/usr/bin/env bash
set -xe

CONF_DIR="$HOME/.jupyter"
mkdir -p $CONF_DIR

cp /jupyter_notebook_config.py $CONF_DIR

jupyter notebook --allow-root "$@"