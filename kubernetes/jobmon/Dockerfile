# docker build  --no-cache -t registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/jobmon -f Dockerfile .
FROM alpine

ENV KUBE_LATEST_VERSION="v1.9.3"

RUN apk add --update ca-certificates \
 && apk add --no-cache bash curl libc6-compat \
 && curl -L http://aliacs-k8s-cn-zhangjiakou.oss-cn-zhangjiakou.aliyuncs.com/public/charts/jobmon -o /usr/local/bin/jobmon \
 && chmod +x /usr/local/bin/jobmon \
 && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl

WORKDIR /root
CMD ["jobmon"]