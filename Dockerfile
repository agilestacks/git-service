FROM golang:1.14-alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache git

WORKDIR /go/src/gits
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
RUN go build github.com/agilestacks/git-service/cmd/gits


FROM alpine:3.11

ENV GIT_API_SECRET "*** unset ***"
ENV HUB_API_SECRET "*** unset ***"
ENV AUTH_API_SECRET "*** unset ***"
ENV HUB_SERVICE_ENDPOINT "*** unset ***"
ENV AUTH_SERVICE_ENDPOINT "*** unset ***"
ENV BUCKET "terraform.agilestacks.com"

ADD https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant /app/gits-key

VOLUME /git

RUN chmod 600 /app/gits-key && \
    apk update && apk upgrade && \
    apk add --no-cache expat git git-subtree tini && \
    apk add --no-cache aws-cli --update-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community && \
    apk upgrade --no-cache python3 --update-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/main && \
    rm /var/cache/apk/*
RUN git config --global user.email "hub@agilestacks.io"
RUN git config --global user.name "Automation Hub"
RUN git config --global uploadpack.allowAnySHA1InWant true

EXPOSE 2022
EXPOSE 8005

WORKDIR /app
COPY --from=builder /go/src/gits/gits ./

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/gits", "-blobs", "s3://$BUCKET/"]
