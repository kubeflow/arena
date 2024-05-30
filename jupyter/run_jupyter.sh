#!/usr/bin/env bash

# Copyright 2018 The Kubeflow Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
