ARG BASE_IMAGE=debian:12-slim

FROM golang:1.23.4 as builder

ARG TARGETOS

ARG TARGETARCH

WORKDIR /workspace

COPY . .

RUN set -eux && \
    VERSION=$(cat VERSION) && \
    make arena-installer OS=${TARGETOS} ARCH=${TARGETARCH} && \
    mv arena-installer-${VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz arena-installer.tar.gz


FROM ${BASE_IMAGE}

ARG TARGETOS

ARG TARGETARCH

WORKDIR /root

RUN apt-get update \
    && apt-get install -y tini \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /workspace/arena-installer.tar.gz .

RUN set -eux && \
    tar -zxvf arena-installer.tar.gz && \
    mv arena-installer-*-${TARGETOS}-${TARGETARCH} arena-installer && \
    arena-installer/install.sh --only-binary && \
    rm -rf arena-installer.tar.gz

COPY entrypoint.sh /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
