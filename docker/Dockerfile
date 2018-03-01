FROM alpine:3.7

ENV USER root

RUN apk add --no-cache curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s http://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && mv /kubectl /bin/kubectl \
    && chmod +x /bin/kubectl

COPY kuberang /
ENTRYPOINT ["/kuberang"]
