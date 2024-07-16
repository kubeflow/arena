FROM ubuntu:22.04

RUN apt update && apt install -y wget

RUN wget https://github.com/kubeflow/arena/releases/download/v0.9.15/arena-installer-0.9.15-f31e1b0-linux-amd64.tar.gz && \
    tar -xvf arena-installer-0.9.15-f31e1b0-linux-amd64.tar.gz && \
    cd arena-installer && \
    bash install.sh --only-binary
