FROM centos:7 as build

RUN yum install -y \
        gcc-c++ \
        ca-certificates \
        wget \
        git \
        make && \
    rm -rf /var/cache/yum/*

ENV GOLANG_VERSION 1.10.3
RUN wget -nv -O - https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p /go/src/github.com/kubeflow/arena

WORKDIR /go/src/github.com/kubeflow/arena
COPY . .

RUN make

RUN wget https://storage.googleapis.com/kubernetes-helm/helm-v2.8.2-linux-amd64.tar.gz && \
    tar -xvf helm-v2.8.2-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/helm && \
    chmod u+x /usr/local/bin/helm

ENV K8S_VERSION v1.11.2
RUN curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl


FROM centos:7

COPY --from=build /go/src/github.com/kubeflow/arena/bin/arena /usr/local/bin/arena

COPY --from=build /usr/local/bin/helm /usr/local/bin/helm

COPY --from=build /go/src/github.com/kubeflow/arena/kubernetes-artifacts /root

COPY --from=build /usr/local/bin/kubectl /usr/local/bin/kubectl

COPY --from=build /go/src/github.com/kubeflow/arena/charts /

ADD run_arena.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/run_arena.sh

ENTRYPOINT ["/usr/local/bin/run_arena.sh"]
    