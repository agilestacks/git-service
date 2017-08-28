FROM alpine

ENV GIT_API_SECRET "*** unset ***"
ENV HUB_API_SECRET "*** unset ***"
ENV AUTH_API_SECRET "*** unset ***"
ENV HUB_SERVICE_ENDPOINT "*** unset ***"
ENV AUTH_SERVICE_ENDPOINT "*** unset ***"
ENV BUCKET "terraform.agilestacks.com"

COPY bin/linux/gits-musl /bin/gits
ADD  https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant /gits-key

VOLUME /git

RUN chmod +x /bin/gits && \
    chmod 600 /gits-key && \
    apk add --update git && \
    rm /var/cache/apk/*

EXPOSE 2022
EXPOSE 8005

ENTRYPOINT gits
CMD ["-blobs", "s3://$BUCKET/"]
