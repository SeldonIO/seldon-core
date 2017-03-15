FROM alpine

MAINTAINER Gurminder Sunner <gs@seldon.io>

ARG KUBECTL_VERSION

RUN apk add --update ca-certificates \
    && apk add --update -t deps curl \
    && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && apk del --purge deps \
    && rm /var/cache/apk/*

EXPOSE 8001

CMD ["/usr/local/bin/kubectl", "proxy", "-p", "8001"]

