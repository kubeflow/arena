#
# Copyright 2026 The Kubeflow Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# 	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG TENSORFLOW_IMAGE=docker.io/tensorflow/tensorflow:2.15.0-gpu

FROM docker.io/library/alpine:3.23.2 AS downloader

# Download MNIST dataset.
RUN set -eux && \
    mkdir -p /data && \
    cd /data && \
    wget https://storage.googleapis.com/tensorflow/tf-keras-datasets/mnist.npz && \
    echo "8a61469f7ea1b51cbae51d4f78837e45 mnist.npz" > checksums.md5 && \
    md5sum -c checksums.md5 && \
    rm checksums.md5

FROM ${TENSORFLOW_IMAGE}

COPY --chmod=644 main.py /code/github.com/kubeflow/arena/examples/tensorflow/mnist/main.py

COPY --from=downloader /data/mnist.npz /data/mnist.npz
