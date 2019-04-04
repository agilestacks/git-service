FROM golang:1.11-alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache git

WORKDIR /go/src/gits
RUN go get github.com/kardianos/govendor
COPY src/gits/vendor /go/src/gits/vendor
RUN /go/bin/govendor sync
COPY src/gits /go/src/gits
RUN go build -v gits


FROM alpine:3.9

ENV GIT_API_SECRET "*** unset ***"
ENV HUB_API_SECRET "*** unset ***"
ENV AUTH_API_SECRET "*** unset ***"
ENV HUB_SERVICE_ENDPOINT "*** unset ***"
ENV AUTH_SERVICE_ENDPOINT "*** unset ***"
ENV BUCKET "terraform.agilestacks.com"

ADD https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant /gits-key

VOLUME /git

RUN chmod 600 /gits-key && \
    apk update && apk upgrade && \
    apk add --no-cache expat git git-subtree tini && \
    apk add --no-cache aws-cli --update-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/testing && \
    rm /var/cache/apk/*
RUN git config --global user.email "hub@agilestacks.io"
RUN git config --global user.name "Automation Hub"
RUN git config --global uploadpack.allowAnySHA1InWant true

EXPOSE 2022
EXPOSE 8005

COPY --from=builder /go/src/gits/gits /bin/gits

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["gits", "-blobs", "s3://$BUCKET/"]
