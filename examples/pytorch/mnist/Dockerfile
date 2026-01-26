#
# Copyright 2024 The Kubeflow Authors.
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

ARG PYTORCH_IMAGE=pytorch/pytorch:2.5.1-cuda12.4-cudnn9-runtime

FROM docker.io/library/alpine:3.23.2 AS downloader

# Download MNIST dataset.
RUN set -eux && \
    mkdir -p /data/MNIST/raw && \
    cd /data/MNIST/raw && \
    files="train-images-idx3-ubyte.gz train-labels-idx1-ubyte.gz t10k-images-idx3-ubyte.gz t10k-labels-idx1-ubyte.gz" && \
    for file in ${files}; do \
        wget https://ossci-datasets.s3.amazonaws.com/mnist/${file}; \
    done && \
    echo "f68b3c2dcbeaaa9fbdd348bbdeb94873 train-images-idx3-ubyte.gz" > checksums.md5 && \
    echo "d53e105ee54ea40749a09fcbcd1e9432 train-labels-idx1-ubyte.gz" >> checksums.md5 && \
    echo "9fb629c4189551a2d022fa330f9573f3 t10k-images-idx3-ubyte.gz" >> checksums.md5 && \
    echo "ec29112dd5afa0611ce80d1b7f02629c t10k-labels-idx1-ubyte.gz" >> checksums.md5 && \
    md5sum -c checksums.md5 && \
    rm checksums.md5 && \
    gzip -d /data/MNIST/raw/*.gz

FROM ${PYTORCH_IMAGE}

COPY requirements.txt .

RUN set -eux && \
    pip install -r requirements.txt && \
    rm requirements.txt

COPY --chmod=644 main.py /code/github.com/kubeflow/arena/examples/pytorch/mnist/main.py

COPY --from=downloader /data /data
