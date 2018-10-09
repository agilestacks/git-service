FROM golang:1.10-alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache git

WORKDIR /go/src/gits
COPY src/gits /go/src/gits

RUN go get github.com/kardianos/govendor
RUN /go/bin/govendor sync
RUN go build -v gits
RUN ls /go/src/gits


FROM alpine:3.8

ARG GIT_VERSION="2.19.1-r0"

ENV GIT_API_SECRET "*** unset ***"
ENV HUB_API_SECRET "*** unset ***"
ENV AUTH_API_SECRET "*** unset ***"
ENV HUB_SERVICE_ENDPOINT "*** unset ***"
ENV AUTH_SERVICE_ENDPOINT "*** unset ***"
ENV BUCKET "terraform.agilestacks.com"

COPY --from=builder /go/src/gits/gits /bin/gits

ADD https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant /gits-key

VOLUME /git

RUN chmod +x /bin/gits && \
    chmod 600 /gits-key && \
    apk update && apk upgrade && \
    apk add --no-cache expat && \
    apk add --no-cache git=${GIT_VERSION} git-subtree=${GIT_VERSION} --update-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/main && \
    apk add --no-cache aws-cli --update-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing && \
    apk add --no-cache tini && \
    rm /var/cache/apk/*
RUN git config --global user.email "hub@agilestack.com"
RUN git config --global user.name "Automation Hub"

EXPOSE 2022
EXPOSE 8005

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["gits", "-blobs", "s3://$BUCKET/"]
